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
	"math"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

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
		return ctrl.Result{}, err
	}

	// if cr is marked for deletion, handle delete and remove finalizer
	if cr.GetDeletionTimestamp() != nil {
		if errDeleting := cloudDBInterface.DeleteDBCluster(cr); errDeleting != nil {
			return ctrl.Result{}, errDeleting
		}
		if errDeletingFinalizer := RemoveFinalizer(dbClusterFinalizer, r.Client, cr); errDeletingFinalizer != nil {
			r.Log.Error(errDeletingFinalizer, "Failed to remove finalizer")
			return ctrl.Result{}, errDeletingFinalizer
		}
		// clean up successful, do not requeue
		r.Log.Info(fmt.Sprintf("Successfully cleaned up dbcluster for %v/%v", cr.GetNamespace(), cr.GetName()))
		return ctrl.Result{}, nil
	}

	clusterExists, errCheckingExistence := cloudDBInterface.DBClusterExists(cr)
	if errCheckingExistence != nil {
		return ctrl.Result{}, errCheckingExistence
	}

	// get masterPassword
	// TODO: we have to store the password somewhere to compare
	// 	if the user changed it or not vs what we created dbCluster with
	passwordSecretName := cr.Spec.MasterUserPasswordSecretRef.SecretRef.Name
	passwordSecretNs := cr.Spec.MasterUserPasswordSecretRef.SecretRef.Namespace
	passwordKey := cr.Spec.MasterUserPasswordSecretRef.PasswordKey
	passwordSecret, errGettingPasswordSecret := getSecret(passwordSecretName, passwordSecretNs, r.Client)
	if errGettingPasswordSecret != nil {
		return ctrl.Result{}, errGettingPasswordSecret
	}
	password, keyFound := passwordSecret.Data[passwordKey]
	if !keyFound {
		return ctrl.Result{}, ErrPasswordKeyNotFound{Message: fmt.Sprintf("%v/%v secret does not contain password key for %v/%v dbcluster",
			passwordSecretNs, passwordSecretName, cr.GetNamespace(), cr.GetName())}
	}
	strPassword := string(password)

	if !clusterExists {
		errCreatingCluster := cloudDBInterface.CreateDBCluster(cr, strPassword)
		if errCreatingCluster != nil {
			return ctrl.Result{}, errCreatingCluster
		}
	}

	clusterUpToDate, errChecking := cloudDBInterface.IsDBClusterUpToDate(cr)
	if errChecking != nil {
		return ctrl.Result{}, errChecking
	}

	if !clusterUpToDate {
		return ctrl.Result{}, cloudDBInterface.ModifyDBCluster(cr, strPassword)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agillappsdboperatorv1alpha1.DBCluster{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: math.MaxInt32}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				oldGen := event.ObjectOld.GetGeneration()
				newGen := event.ObjectNew.GetGeneration()
				return oldGen == newGen
			},
		}).
		Complete(r)
}
