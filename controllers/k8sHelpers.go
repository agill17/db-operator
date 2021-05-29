package controllers

import (
	"context"
	"fmt"
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

func getSecretValue(name, namespace, key string, client client.Client) (string, error) {
	s, err := getSecret(name, namespace, client)
	if err != nil {
		return "", err
	}

	val, ok := s.Data[key]
	if !ok {
		return "", ErrSecretMissingKey{
			Message: fmt.Sprintf("%v/%v does not contain %v key", namespace, name, key)}
	}
	return string(val), nil
}
