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
	"github.com/agill17/db-operator/factory"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

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

//+kubebuilder:rbac:groups=agill.apps.db-operator.agill.apps.db-operator,resources=dbclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=agill.apps.db-operator.agill.apps.db-operator,resources=dbclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=agill.apps.db-operator.agill.apps.db-operator,resources=dbclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DBCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
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

	providerCr := &agillappsdboperatorv1alpha1.Provider{}
	if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: cr.Spec.ProviderRef}, providerCr); err != nil {
		return ctrl.Result{}, err
	}

	//TODO: add finalizer here

	cloudDBInterface, err := factory.NewCloudDB(providerCr, cr.Spec.Region)
	if err != nil {
		return ctrl.Result{}, err
	}

	clusterExists, errCheckingExistence := cloudDBInterface.DBClusterExists()
	if errCheckingExistence != nil {
		return ctrl.Result{}, errCheckingExistence
	}

	if cr.GetDeletionTimestamp() != nil {
		if clusterExists {
			if errDeleting := cloudDBInterface.DeleteDBCluster(); errDeleting != nil {
				return ctrl.Result{}, errDeleting
			}
			// TODO: remove finalizer here
			return ctrl.Result{}, nil
		}
	}

	if !clusterExists {
		errCreatingCluster := cloudDBInterface.CreateDBCluster()
		if errCreatingCluster != nil {
			return ctrl.Result{}, errCreatingCluster
		}
	}

	clusterUpToDate, errChecking := cloudDBInterface.IsDBClusterUpToDate()
	if errChecking != nil {
		return ctrl.Result{}, errChecking
	}

	if !clusterUpToDate {
		return ctrl.Result{}, cloudDBInterface.ModifyDBCluster()
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DBClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agillappsdboperatorv1alpha1.DBCluster{}).
		Complete(r)
}
