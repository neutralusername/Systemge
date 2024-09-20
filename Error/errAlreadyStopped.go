package Error

import "errors"

type ErrAlreadyStopped struct {
	err error
}

func (e *ErrAlreadyStopped) Error() string {
	return e.err.Error()
}

func NewErrAlreadyStopped(err string) *ErrAlreadyStopped {
	if err == "" {
		err = "already stopped"
	}
	return &ErrAlreadyStopped{
		err: errors.New(err),
	}
}
