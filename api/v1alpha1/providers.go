package v1alpha1

import v1 "k8s.io/api/core/v1"

type ProviderType string

const (
	AWS   ProviderType = "aws"
	GCP   ProviderType = "gcp"
	Azure ProviderType = "azure"
)

type Provider struct {
	// +kubebuilder:validation:Enum=aws;gcp;azure
	Type      ProviderType       `json:"type,required"`
	SecretRef v1.SecretReference `json:"secretRef,required"`
}
