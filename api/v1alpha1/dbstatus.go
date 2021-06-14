package v1alpha1

// used by factories to return when a describe DB per cloud is called
type DBStatus struct {
	Exists       bool
	CurrentPhase string
	Endpoint     string
}
