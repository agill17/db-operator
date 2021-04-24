package vault

type ErrProviderMissingVaultAuthBackendPath struct {
	Message string
}

func (e ErrProviderMissingVaultAuthBackendPath) Error() string {
	return e.Message
}

type ErrProviderMissingAuthBackendRole struct {
	Message string
}

func (e ErrProviderMissingAuthBackendRole) Error() string {
	return e.Message
}