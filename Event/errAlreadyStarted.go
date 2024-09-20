package Event

type ErrAlreadyStarted string

func (e ErrAlreadyStarted) Error() Type {
	return Type(e)
}
