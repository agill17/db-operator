package v1alpha1

type Phase string

const (
	Creating  Phase = "creating"
	Updating  Phase = "updating"
	Deleting  Phase = "deleting"
	Available Phase = "available"
)
