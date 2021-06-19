package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func (r *DBInstanceReconciler) dbInstancePredicates() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(event event.UpdateEvent) bool {
			return event.ObjectOld.GetGeneration() != event.ObjectNew.GetGeneration()
		},
	}
}
