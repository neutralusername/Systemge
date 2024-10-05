package Tools

import (
	"errors"
	"sync"
	"time"
)

var ErrAlreadyTriggered = errors.New("timeout already triggered")

const (
	cancelTimeout = iota
	refreshTimeout
	triggerTimeout
)

type Timeout struct {
	durationMs         uint64
	onTrigger          func()
	interactionChannel chan uint64
	triggered          bool
	mutex              sync.Mutex
	mayBeCancelled     bool
	triggeredChannel   chan struct{}
}

// duration 0 == must be triggered manually
func NewTimeout(durationMs uint64, onTrigger func(), mayBeCancelled bool) *Timeout {
	timeout := &Timeout{
		durationMs:         durationMs,
		onTrigger:          onTrigger,
		triggered:          false,
		interactionChannel: make(chan uint64),
		mayBeCancelled:     mayBeCancelled,
		triggeredChannel:   make(chan struct{}),
	}
	go timeout.handleTrigger()
	return timeout
}

func (timeout *Timeout) handleTrigger() {
	for {
		var timeoutChannel <-chan time.Time
		if timeout.durationMs > 0 {
			timeoutChannel = time.After(time.Duration(timeout.durationMs) * time.Millisecond)
		}
		select {
		case val := <-timeout.interactionChannel:
			switch val {
			case triggerTimeout:
				timeout.onTrigger()
				close(timeout.triggeredChannel)
				return
			case refreshTimeout:
			case cancelTimeout:
				close(timeout.triggeredChannel)
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
	return timeout.durationMs
}

func (timeout *Timeout) SetDuration(duration uint64) {
	timeout.durationMs = duration
}

func (timeout *Timeout) IsTriggered() bool {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()
	return timeout.triggered
}

func (timeout *Timeout) GetTriggeredChannel() <-chan struct{} {
	return timeout.triggeredChannel
}

func (timeout *Timeout) Trigger() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()
	if timeout.triggered {
		return ErrAlreadyTriggered
	}
	timeout.triggered = true
	timeout.interactionChannel <- triggerTimeout
	return nil
}

func (timeout *Timeout) Refresh() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()
	if timeout.triggered {
		return ErrAlreadyTriggered
	}
	timeout.interactionChannel <- refreshTimeout
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
	timeout.interactionChannel <- cancelTimeout
	return nil
}
