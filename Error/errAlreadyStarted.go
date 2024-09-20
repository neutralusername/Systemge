package Error

import "errors"

type ErrAlreadyStarted struct {
	err error
}

func (e *ErrAlreadyStarted) Error() string {
	return e.err.Error()
}

func NewErrAlreadyStarted(err string) *ErrAlreadyStarted {
	if err == "" {
		err = "already started"
	}
	return &ErrAlreadyStarted{
		err: errors.New(err),
	}
}
