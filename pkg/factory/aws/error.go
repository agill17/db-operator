package aws

type ErrorProviderMissingAwsAccessKeyID struct {
	Message string
}

func (e ErrorProviderMissingAwsAccessKeyID) Error() string {
	return e.Message
}

type ErrorProviderMissingAwsSecretAccessKey struct {
	Message string
}

func (e ErrorProviderMissingAwsSecretAccessKey) Error() string {
	return e.Message
}

type ErrRequeueNeeded struct {
	Message string
}

func (e ErrRequeueNeeded) Error() string {
	return e.Message
}

type ErrDBClusterPasswordKeyInSecretNotFound struct {
	Message string
}

func (e ErrDBClusterPasswordKeyInSecretNotFound) Error() string {
	return e.Message
}

type ErrDBClusterDeletionProtectionEnabled struct {
	Message string
}

func (e ErrDBClusterDeletionProtectionEnabled) Error() string {
	return e.Message
}

type ErrInvalidTypeWasPassedIn struct {
	Message string
}

func (e ErrInvalidTypeWasPassedIn) Error() string {
	return e.Message
}
