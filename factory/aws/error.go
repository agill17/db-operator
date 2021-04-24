package aws

type ErrorProviderMissingAwsAccessKeyID struct {
	Message string
}
func (e ErrorProviderMissingAwsAccessKeyID) Error() string{
	return e.Message
}


type ErrorProviderMissingAwsSecretAccessKey struct {
	Message string
}
func (e ErrorProviderMissingAwsSecretAccessKey) Error() string{
	return e.Message
}