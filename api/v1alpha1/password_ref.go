package v1alpha1

import v1 "k8s.io/api/core/v1"

type PasswordRef struct {
	SecretRef   *v1.LocalObjectReference `json:"secretRef,required"`
	PasswordKey string                   `json:"passwordKey,required"`
}
