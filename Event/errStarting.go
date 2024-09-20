package Event

type EventStarting string

func (e eventStarting) Error() Type {
	return Type(e)
}
