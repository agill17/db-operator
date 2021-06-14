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
	"github.com/agill17/db-operator/pkg/factory"
	"github.com/agill17/db-operator/pkg/factory/aws"
	"github.com/agill17/db-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/agill17/db-operator/api/v1alpha1"
)

// DBClusterReconciler reconciles a DBCluster object
type DBClusterReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	CloudDBInterface factory.CloudDB
}

var (
	groupName          = v1alpha1.GroupVersion.Group
	groupVersion       = v1alpha1.GroupVersion.Version
	dbClusterFinalizer = fmt.Sprintf("%s/%s-finalizer", groupName, groupVersion)
)

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DBClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	namespacedName := req.NamespacedName.String()
	_ = r.Log.WithValues("dbcluster", req.NamespacedName)
	cr := &v1alpha1.DBCluster{}
	if err := r.Client.Get(context.TODO(), req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// add finalizer if needed
	if errAddingFinalizer := utils.AddFinalizer(dbClusterFinalizer, r.Client, cr); errAddingFinalizer != nil {
		r.Log.Error(errAddingFinalizer, "Failed to add finalizer")
		return ctrl.Result{}, errAddingFinalizer
	}

	// get provider secret
	providerSecret, errGettingSecret := utils.GetSecret(cr.Spec.Provider.SecretRef.Name,
		cr.Spec.Provider.SecretRef.Namespace, r.Client)
	if errGettingSecret != nil {
		r.Log.Error(errGettingSecret, "Failed to get provider secret")
		return ctrl.Result{}, errGettingSecret
	}

	// setup cloud clients
	if r.CloudDBInterface == nil {
		cloudDBInterface, err := factory.NewCloudDB(r.Log, cr.Spec.Provider.Type, providerSecret, cr.Spec.Region)
		if err != nil {
			r.Log.Error(err, "Failed to create a NewCloudDB client interface")
			return ctrl.Result{}, err
		}
		r.CloudDBInterface = cloudDBInterface
	}

	dbStatus, errCheckingExistence := r.CloudDBInterface.DBClusterExists(cr.GetDBClusterID())
	if errCheckingExistence != nil {
		return ctrl.Result{}, errCheckingExistence
	}
	if dbStatus.Exists && dbStatus.CurrentPhase != string(v1alpha1.Available) {
		r.Log.Info(fmt.Sprintf("%v - DBCluster exists, but is not yet ready. Current status: %v", namespacedName, dbStatus.CurrentPhase))
		return ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
	}

	// if cr is marked for deletion, handle delete and remove finalizer
	if cr.GetDeletionTimestamp() != nil {
		r.Log.Info(fmt.Sprintf("%v - is marked for deletion", namespacedName))
		if errUpdatingPhase := utils.UpdateStatusPhase(
			v1alpha1.Deleting, cr, r.Client); errUpdatingPhase != nil {
			return ctrl.Result{}, errUpdatingPhase
		}
		if dbStatus.Exists {
			if errDeleting := r.CloudDBInterface.DeleteDBCluster(cr); errDeleting != nil {
				if _, ok := errDeleting.(aws.ErrRequeueNeeded); ok {
					return ctrl.Result{Requeue: true}, nil
				}
				return ctrl.Result{}, errDeleting
			}
		}
		if errDeletingFinalizer := utils.RemoveFinalizer(dbClusterFinalizer, r.Client, cr); errDeletingFinalizer != nil {
			r.Log.Error(errDeletingFinalizer, "Failed to remove finalizer")
			return ctrl.Result{}, errDeletingFinalizer
		}
		r.Log.Info(fmt.Sprintf("%v - deleted successfully", namespacedName))
		return ctrl.Result{}, nil
	}

	// get masterPassword
	// TODO: Generate password and make masterUserPassword optional
	passSecretName := cr.Spec.PasswordRef.SecretRef.Name
	passSecretNs := cr.GetNamespace()
	passwordKey := cr.Spec.PasswordRef.PasswordKey
	dbPass, _, errFetchingKey := utils.GetSecretValue(passSecretName,
		passSecretNs, passwordKey, r.Client)
	if errFetchingKey != nil {
		return ctrl.Result{}, errFetchingKey
	}

	if !dbStatus.Exists {
		r.Log.Info(fmt.Sprintf("%v - does not exist in cloud, creating now", namespacedName))
		if errCreatingDBCluster := r.CloudDBInterface.CreateDBCluster(cr, dbPass); errCreatingDBCluster != nil {
			r.Log.Error(errCreatingDBCluster, fmt.Sprintf("%v - failed to create dbcluster", namespacedName))
			return ctrl.Result{}, errCreatingDBCluster
		}
		return ctrl.Result{Requeue: true}, utils.UpdateStatusPhase(v1alpha1.Creating, cr, r.Client)
	}

	isUpToDate, modifyIn, errChecking := r.CloudDBInterface.IsDBClusterUpToDate(cr)
	if errChecking != nil {
		return ctrl.Result{}, errChecking
	}

	if !isUpToDate {
		errUpdating := r.CloudDBInterface.ModifyDBCluster(modifyIn)
		if errUpdating != nil {
			return ctrl.Result{}, errUpdating
		}
		r.Log.Info(fmt.Sprintf("%v - updating", namespacedName))
		return ctrl.Result{Requeue: true}, utils.UpdateStatusPhase(v1alpha1.Updating, cr, r.Client)
	}

	svcResult, svcName, errReconcilingSvc := r.createOrUpdateExternalNameSvc(cr, dbStatus.Endpoint)
	if errReconcilingSvc != nil {
		return ctrl.Result{}, errReconcilingSvc
	}
	r.Log.Info(fmt.Sprintf("%s - ExternalName service %s", svcName, svcResult))

	if err := utils.UpdateStatusPhase(v1alpha1.Available, cr, r.Client); err != nil {
		return ctrl.Result{}, err
	}
	r.Log.Info(fmt.Sprintf("%v - reconciled", namespacedName))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBCluster{}, builder.WithPredicates(r.dbClusterPredicates())).
		Owns(&v1.Service{}).
		Watches(
			&source.Kind{Type: &v1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.dbClusterSecretsEventHandlerFunc()),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Complete(r)
}

func (r *DBClusterReconciler) createOrUpdateExternalNameSvc(cr *v1alpha1.DBCluster, endpoint string) (string, string, error) {
	svcName := cr.GetName()
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: cr.GetNamespace(),
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: endpoint,
		},
	}

	res, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, svc, func() error {
		return controllerutil.SetControllerReference(cr, svc, r.Scheme)
	})
	return string(res), svcName, err
}
