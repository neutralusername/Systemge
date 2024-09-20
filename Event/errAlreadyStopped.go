package Event

type ErrAlreadyStopped string

func (e ErrAlreadyStopped) Error() Type {
	return Type(e)
}
