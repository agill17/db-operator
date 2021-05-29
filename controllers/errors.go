package controllers

type ErrSecretMissingKey struct {
	Message string
}

func (e ErrSecretMissingKey) Error() string {
	return e.Message
}
