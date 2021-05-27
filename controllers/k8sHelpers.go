package controllers

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getSecret(name, namespace string, client client.Client) (*v1.Secret, error) {
	secret := &v1.Secret{}
	if err := client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, secret); err != nil {
		return nil, err
	}
	return secret, nil
}
