package utils

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSecret(name, namespace string, client client.Client) (*v1.Secret, error) {
	secret := &v1.Secret{}
	if err := client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, secret); err != nil {
		return nil, err
	}
	return secret, nil
}

func GetSecretValue(name, namespace, key string, client client.Client) (string, string, error) {
	s, err := GetSecret(name, namespace, client)
	if err != nil {
		return "", "", err
	}

	val, ok := s.Data[key]
	if !ok {
		return "", "", ErrSecretMissingKey{
			Message: fmt.Sprintf("%v/%v does not contain %v key", namespace, name, key)}
	}
	return string(val), s.GetResourceVersion(), nil
}
