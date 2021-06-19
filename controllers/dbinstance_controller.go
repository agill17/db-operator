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
	"github.com/agill17/db-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/agill17/db-operator/api/v1alpha1"
)

// DBInstanceReconciler reconciles a DBInstance object
type DBInstanceReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	CloudDBInterface factory.CloudDB
}

var (
	dbInstanceFinalizer = fmt.Sprintf("%s/%s-dbinstance", groupName, groupVersion)
)

//+kubebuilder:rbac:groups=agill.apps.db-operator,resources=dbinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=agill.apps.db-operator,resources=dbinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=agill.apps.db-operator,resources=dbinstances/finalizers,verbs=update
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DBInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("dbinstance", req.NamespacedName)
	namespacedName := req.NamespacedName.String()
	cr := &v1alpha1.DBInstance{}
	if errGettingCr := r.Client.Get(context.TODO(), req.NamespacedName, cr); errGettingCr != nil {
		if errors.IsNotFound(errGettingCr) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errGettingCr
	}

	// add finalizer
	if errAddingFinalizer := utils.AddFinalizer(dbInstanceFinalizer, r.Client, cr); errAddingFinalizer != nil {
		return ctrl.Result{}, errAddingFinalizer
	}

	// get provider secret
	providerSecret, errGettngSecret := utils.GetSecret(cr.Spec.Provider.SecretRef.Name, cr.Spec.Provider.SecretRef.Namespace, r.Client)
	if errGettngSecret != nil {
		return ctrl.Result{}, errGettngSecret
	}

	// create cloud client(s)
	if r.CloudDBInterface == nil {
		cloudDBInterface, err := factory.NewCloudDB(r.Log, cr.Spec.Provider.Type, providerSecret, cr.Spec.Region)
		if err != nil {
			return ctrl.Result{}, err
		}
		r.CloudDBInterface = cloudDBInterface
	}

	// get instance status
	instanceStatus, err := r.CloudDBInterface.DBInstanceExists(cr)
	if err != nil {
		return ctrl.Result{}, err
	}

	if instanceStatus.Exists && instanceStatus.CurrentPhase != string(v1alpha1.Available) {
		r.Log.Info(fmt.Sprintf("%s - exists but not yet available. Current status: %s", namespacedName, instanceStatus.CurrentPhase))
		return ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
	}

	// handle delete
	if cr.GetDeletionTimestamp() != nil {
		r.Log.Info(fmt.Sprintf("%v - is marked for deletion", namespacedName))

		if instanceStatus.Exists {
			// if part of dbcluster, wait for dbcluster to delete first
			hasDBClusterFinalizer, _ := utils.ListContainsString(cr.GetFinalizers(), dbClusterFinalizer)
			if hasDBClusterFinalizer {
				r.Log.Info(fmt.Sprintf("%s - is part of DBCluster, waiting for dbcluster to get deleted first", namespacedName))
				return ctrl.Result{RequeueAfter: 30 * time.Second, Requeue: true}, nil
			}
			errDeleting := r.CloudDBInterface.DeleteDBInstance(cr)
			if errDeleting != nil {
				return ctrl.Result{}, errDeleting
			}
		}
		if errRemovingFinalizer := utils.RemoveFinalizer(dbInstanceFinalizer, r.Client, cr); errRemovingFinalizer != nil {
			return ctrl.Result{}, errRemovingFinalizer
		}
		r.Log.Info(fmt.Sprintf("%v - deleted successfully", namespacedName))
		return ctrl.Result{}, nil
	}

	// get password
	insPass := ""
	if cr.Spec.DBClusterID == "" {
		secretValue, _, err := utils.GetSecretValue(cr.Spec.PasswordRef.SecretRef.Name, cr.GetNamespace(), cr.Spec.PasswordRef.PasswordKey, r.Client)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("%s - could not get instance password from secret", namespacedName))
			return ctrl.Result{}, err
		}
		insPass = secretValue
	}

	if !instanceStatus.Exists {
		errCreating := r.CloudDBInterface.CreateDBInstance(cr, insPass)
		if errCreating != nil {
			return ctrl.Result{}, errCreating
		}
		r.Log.Info(fmt.Sprintf("%s - instance does not exist, creating now.", namespacedName))
		return ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}, utils.UpdateStatusPhase(v1alpha1.Creating, cr, r.Client)
	}

	// check if update needed and modify as needed

	// create external name service
	svcResult, svcName, errReconcilingSvc := createOrUpdateExternalNameSvc(cr, instanceStatus.Endpoint, r.Client, r.Scheme)
	if errReconcilingSvc != nil {
		return ctrl.Result{}, errReconcilingSvc
	}

	errUpdatingStatus := utils.UpdateStatusPhase(v1alpha1.Available, cr, r.Client)
	if errUpdatingStatus != nil {
		return ctrl.Result{}, errUpdatingStatus
	}
	r.Log.Info(fmt.Sprintf("%s - ExternalName service %s", svcName, svcResult))
	r.Log.Info(fmt.Sprintf("%s - reconciled", namespacedName))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DBInstance{}, builder.WithPredicates(r.dbInstancePredicates())).
		Owns(&v1.Service{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool { return false },
		})).
		Complete(r)
}
