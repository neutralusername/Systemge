package Tools

// timeout runs for a specified duration and then triggers a function.

// may be refreshed to delay the trigger.
// may be cancelled to prevent the trigger.

// may be triggered manually.
// timeout may only be triggered/cancelled once.

type Timeout struct {
	duration  uint64
	onTrigger func()
}

func NewTimeout(duration uint64, onTrigger func()) *Timeout {
	return &Timeout{
		duration: duration,
	}
}

func (timeout *Timeout) GetDuration() uint64 {
	return timeout.duration
}

func (timeout *Timeout) SetDuration(duration uint64) {
	timeout.duration = duration
}

func (timeout *Timeout) Trigger() error {

}

func (timeout *Timeout) Cancel() error {

}

func (timeout *Timeout) Refresh() error {

}
