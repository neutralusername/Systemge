package Event

type ErrAlreadyStopped Type

func (e ErrAlreadyStopped) Error() string {
	return string(e)
}
