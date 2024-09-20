package Event

type ErrAlreadyStarted Type

func (e ErrAlreadyStarted) Error() string {
	return string(e)
}
