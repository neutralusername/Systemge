package Tools

import (
	"errors"
	"sync"
	"time"
)

var ErrAlreadyTriggered = errors.New("timeout already triggered")

type Timeout struct {
	duration           uint64
	onTrigger          func()
	interactionChannel chan uint64
	triggered          bool
	mutex              sync.Mutex
	mayBeCancelled     bool
	triggeredChannel   chan bool
}

// duration 0 == must be triggered manually
func NewTimeout(duration uint64, onTrigger func(), mayBeCancelled bool) *Timeout {
	timeout := &Timeout{
		duration:           duration,
		onTrigger:          onTrigger,
		triggered:          false,
		interactionChannel: make(chan uint64),
		mayBeCancelled:     mayBeCancelled,
		triggeredChannel:   make(chan bool),
	}
	go timeout.handleTrigger()
	return timeout
}

func (timeout *Timeout) handleTrigger() {
	for {
		var timeoutChannel <-chan time.Time
		if timeout.duration > 0 {
			timeoutChannel = time.After(time.Duration(timeout.duration))
		}
		select {
		case val := <-timeout.interactionChannel:
			switch val {
			case 0:
				timeout.onTrigger()
				close(timeout.triggeredChannel)
				return
			case 1:
			case 2:
				timeout.onTrigger()
				return
			}
		case <-timeoutChannel:
			timeout.mutex.Lock()
			if timeout.triggered {
				timeout.mutex.Unlock()
				return
			}
			timeout.triggered = true
			timeout.mutex.Unlock()

			timeout.onTrigger()
			close(timeout.triggeredChannel)
			return
		}
	}
}

func (timeout *Timeout) GetDuration() uint64 {
	return timeout.duration
}

func (timeout *Timeout) SetDuration(duration uint64) {
	timeout.duration = duration
}

func (timeout *Timeout) IsTriggered() bool {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()
	return timeout.triggered
}

func (timeout *Timeout) GetTriggeredChannel() <-chan bool {
	return timeout.triggeredChannel
}

func (timeout *Timeout) Trigger() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()
	if timeout.triggered {
		return ErrAlreadyTriggered
	}
	timeout.triggered = true
	timeout.interactionChannel <- 0
	return nil
}

func (timeout *Timeout) Refresh() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()
	if timeout.triggered {
		return ErrAlreadyTriggered
	}
	timeout.interactionChannel <- 1
	return nil
}

func (timeout *Timeout) Cancel() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()
	if timeout.triggered {
		return ErrAlreadyTriggered
	}
	if !timeout.mayBeCancelled {
		return errors.New("timeout cannot be cancelled")
	}
	timeout.triggered = true
	timeout.interactionChannel <- 2
	return nil
}
