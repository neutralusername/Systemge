package Event

type EventStarting Type

func (e EventStarting) Error() string {
	return string(e)
}
