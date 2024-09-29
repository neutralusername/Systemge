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
}

func NewTimeout(duration uint64, onTrigger func()) *Timeout {
	timeout := &Timeout{
		duration:           duration,
		onTrigger:          onTrigger,
		triggered:          false,
		interactionChannel: make(chan uint64),
	}
	go timeout.handleTrigger()
	return timeout
}

func (timeout *Timeout) handleTrigger() {
	for {
		select {
		case val := <-timeout.interactionChannel:
			switch val {
			case 0:
				timeout.onTrigger()
				return
			case 1:
			case 2:
				return
			}
		case <-time.After(time.Duration(timeout.duration)):
			timeout.triggered = true
			timeout.onTrigger()
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
	timeout.triggered = true
	timeout.interactionChannel <- 2
	return nil
}
