package controllers

type ErrPasswordKeyNotFound struct {
	Message string
}

func (e ErrPasswordKeyNotFound) Error() string {
	return e.Message
}
