package utils

import (
	"context"
	"github.com/agill17/db-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateStatusPhase(phase v1alpha1.Phase, object client.Object, client client.Client) error {
	// i hate this
	switch o := object.(type) {
	case *v1alpha1.DBCluster:
		if o.Status.Phase != phase {
			o.Status.Phase = phase
			return client.Status().Update(context.TODO(), o)
		}
	case *v1alpha1.DBInstance:
		if o.Status.Phase != phase {
			o.Status.Phase = phase
			return client.Status().Update(context.TODO(), o)
		}
	}
	return nil

}
