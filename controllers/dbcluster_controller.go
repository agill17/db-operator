/*
Copyright 2021 agill17.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/agill17/db-operator/controllers/factory"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agillappsdboperatorv1alpha1 "github.com/agill17/db-operator/api/v1alpha1"
)

// DBClusterReconciler reconciles a DBCluster object
type DBClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var (
	groupName          = agillappsdboperatorv1alpha1.GroupVersion.Group
	groupVersion       = agillappsdboperatorv1alpha1.GroupVersion.Version
	dbClusterFinalizer = fmt.Sprintf("%s/%s-finalizer", groupName, groupVersion)
)

//+kubebuilder:rbac:groups=agill.apps.db-operator,resources=dbclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=agill.apps.db-operator,resources=dbclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=agill.apps.db-operator,resources=dbclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *DBClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("dbcluster", req.NamespacedName)

	cr := &agillappsdboperatorv1alpha1.DBCluster{}
	if err := r.Client.Get(context.TODO(), req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// add finalizer if needed
	if errAddingFinalizer := AddFinalizer(dbClusterFinalizer, r.Client, cr); errAddingFinalizer != nil {
		r.Log.Error(errAddingFinalizer, "Failed to add finalizer")
		return ctrl.Result{}, errAddingFinalizer
	}

	// get provider secret
	providerSecret, errGettingSecret := getSecret(cr.Spec.Provider.SecretRef.Name,
		cr.Spec.Provider.SecretRef.Namespace, r.Client)
	if errGettingSecret != nil {
		r.Log.Error(errGettingSecret, "Failed to get provider secret")
		return ctrl.Result{}, errGettingSecret
	}

	// setup cloud clients
	cloudDBInterface, err := factory.NewCloudDB(cr.Spec.Provider.Type, providerSecret, cr.Spec.Region)
	if err != nil {
		r.Log.Error(err, "Failed to create a NewCloudDB client interface")
		return ctrl.Result{}, err
	}

	clusterExists, status, errCheckingExistence := cloudDBInterface.DBClusterExists(cr.GetDBClusterID())
	if errCheckingExistence != nil {
		return ctrl.Result{}, errCheckingExistence
	}
	if clusterExists && status != "available" {
		r.Log.Info(fmt.Sprintf("%v - DBCluster exists, but is not yet ready. Current status: %v", req.NamespacedName.String(), status))
		return ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
	}

	// if cr is marked for deletion, handle delete and remove finalizer
	if cr.GetDeletionTimestamp() != nil {
		r.Log.Info(fmt.Sprintf("%v - is marked for deletion", req.NamespacedName.String()))
		if errUpdatingPhase := updateStatusPhase(
			agillappsdboperatorv1alpha1.ClusterDeleting, cr, r.Client); errUpdatingPhase != nil {
			return ctrl.Result{}, errUpdatingPhase
		}
		if clusterExists {
			if errDeleting := cloudDBInterface.DeleteDBCluster(cr); errDeleting != nil {
				return ctrl.Result{}, errDeleting
			}
		}

		if errDeletingFinalizer := RemoveFinalizer(dbClusterFinalizer, r.Client, cr); errDeletingFinalizer != nil {
			r.Log.Error(errDeletingFinalizer, "Failed to remove finalizer")
			return ctrl.Result{}, errDeletingFinalizer
		}
		// clean up successful, do not requeue
		r.Log.Info(fmt.Sprintf("Successfully cleaned up dbcluster for %v/%v", cr.GetNamespace(), cr.GetName()))
		return ctrl.Result{}, nil
	}

	// get masterPassword
	// TODO: we have to store the password somewhere to compare
	// 	if the user changed it or not vs what we created dbCluster with
	passSecretName := cr.Spec.MasterUserPasswordSecretRef.SecretRef.Name
	passSecretNs := cr.Spec.MasterUserPasswordSecretRef.SecretRef.Namespace
	passwordKey := cr.Spec.MasterUserPasswordSecretRef.PasswordKey
	dbPass, errFetchingKey := getSecretValue(passSecretName,
		passSecretNs, passwordKey, r.Client)
	if errFetchingKey != nil {
		return ctrl.Result{}, errFetchingKey
	}

	if !clusterExists {
		if errUpdatingPhase := updateStatusPhase(
			agillappsdboperatorv1alpha1.ClusterCreating, cr, r.Client); errUpdatingPhase != nil {
			return ctrl.Result{}, errUpdatingPhase
		}
		r.Log.Info(fmt.Sprintf("%v - does not exist in cloud, creating now", req.NamespacedName.String()))
		if errCreatingDBCluster := cloudDBInterface.CreateDBCluster(cr, dbPass); errCreatingDBCluster != nil {
			r.Log.Error(errCreatingDBCluster, fmt.Sprintf("%v - failed to create dbcluster", req.NamespacedName.String()))
			return ctrl.Result{}, errCreatingDBCluster
		}
		return ctrl.Result{Requeue: true}, nil
	}

	isUpToDate, modifyIn, errChecking := cloudDBInterface.IsDBClusterUpToDate(cr)
	if errChecking != nil {
		return ctrl.Result{}, errChecking
	}
	if !isUpToDate {
		errUpdatingStatus := updateStatusPhase(agillappsdboperatorv1alpha1.ClusterUpdating, cr, r.Client)
		if errUpdatingStatus != nil {
			return ctrl.Result{}, errUpdatingStatus
		}
		return ctrl.Result{}, cloudDBInterface.ModifyDBCluster(modifyIn, dbPass)
	}

	if err := updateStatusPhase(agillappsdboperatorv1alpha1.ClusterAvailable, cr, r.Client); err != nil {
		return ctrl.Result{}, err
	}
	r.Log.Info(fmt.Sprintf("%v - reconciled", req.NamespacedName.String()))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agillappsdboperatorv1alpha1.DBCluster{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		//WithEventFilter(predicate.Funcs{
		//	UpdateFunc: func(event event.UpdateEvent) bool {
		//		oldGen := event.ObjectOld.GetGeneration()
		//		newGen := event.ObjectNew.GetGeneration()
		//		return oldGen == newGen
		//	},
		//}).
		Complete(r)
}
