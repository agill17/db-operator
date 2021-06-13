package controllers

import (
	"context"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *DBClusterReconciler) dbClusterPredicates() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return updateEvent.ObjectOld.GetGeneration() != updateEvent.ObjectNew.GetGeneration()
		},
	}
}

func (r *DBClusterReconciler) dbClusterSecretsEventHandlerFunc() func(object client.Object) []reconcile.Request {
	return func(object client.Object) []reconcile.Request {
		s, ok := object.(*v1.Secret)
		if !ok {
			r.Log.Info(fmt.Sprintf("Expected secret to be returned by watch, but got: %T", object))
			return []reconcile.Request{}
		}
		sName := s.GetName()
		sNamespace := s.GetNamespace()
		var result []reconcile.Request
		dbClusterList := &v1alpha1.DBClusterList{}
		errListingDBClusters := r.Client.List(context.TODO(), dbClusterList)
		if errListingDBClusters != nil {
			r.Log.Error(errListingDBClusters, "Failed to parse watch.secret event because we failed to list all dbclusters")
			return result
		}

		for _, dbCluster := range dbClusterList.Items {
			dbClusterProviderSecretName := dbCluster.Spec.Provider.SecretRef.Name
			dbClusterProviderSecretNamespace := dbCluster.Spec.Provider.SecretRef.Namespace
			dbClusterPasswordSecretName := dbCluster.Spec.PasswordRef.SecretRef.Name
			dbClusterPasswordSecretNamespace := dbCluster.GetNamespace()
			if (sName == dbClusterProviderSecretName && sNamespace == dbClusterProviderSecretNamespace) ||
				(sName == dbClusterPasswordSecretName && sNamespace == dbClusterPasswordSecretNamespace) {

				result = append(result, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dbCluster.GetName(),
						Namespace: dbCluster.GetNamespace(),
					},
				})
			}
		}
		return result
	}
}
