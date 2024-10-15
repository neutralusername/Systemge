package Tools

import (
	"errors"
	"sync"
	"time"
)

var ErrAlreadyTriggered = errors.New("timeout already triggered")

type Timeout struct {
	onTrigger func()

	timeoutNs int64

	triggerAt time.Time

	cancellable bool

	interactionChannel chan int64
	triggeredChannel   chan struct{}

	isExpired bool
	mutex     sync.Mutex
}

// timeoutNs 0 == must be triggered manually
func NewTimeout(timeoutNs int64, onTrigger func(), cancellable bool) *Timeout {
	timeout := &Timeout{
		timeoutNs:          timeoutNs,
		onTrigger:          onTrigger,
		isExpired:          false,
		interactionChannel: make(chan int64),
		cancellable:        cancellable,
		triggeredChannel:   make(chan struct{}),
	}
	go timeout.handleTrigger()
	return timeout
}

func (timeout *Timeout) handleTrigger() {
	for {
		var timeoutChannel <-chan time.Time

		if timeout.timeoutNs > 0 {
			triggerTimestamp := time.Now().UnixNano() + timeout.timeoutNs
			timeoutChannel = time.After(time.Duration(triggerTimestamp - time.Now().UnixNano()))
			timeout.triggerAt = time.Unix(0, triggerTimestamp)
		} else {
			timeout.triggerAt = time.Time{}
		}

		select {
		case val := <-timeout.interactionChannel:
			if val > 0 {
				timeout.timeoutNs = val
			} else {
				timeout.isExpired = true
				return
			}
		case <-timeoutChannel:
			timeout.mutex.Lock()
			if timeout.isExpired {
				timeout.mutex.Unlock()
				return
			}
			timeout.isExpired = true
			timeout.mutex.Unlock()

			timeout.onTrigger()
			close(timeout.triggeredChannel)
			return
		}
	}
}

func (timeout *Timeout) GetTimeoutNs() int64 {
	return timeout.timeoutNs
}

func (timeout *Timeout) TriggerAt() int64 {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	return timeout.triggerAt.UnixNano()
}

func (timeout *Timeout) IsExpired() bool {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	return timeout.isExpired
}

func (timeout *Timeout) GetTriggeredChannel() <-chan struct{} {
	return timeout.triggeredChannel
}

func (timeout *Timeout) Trigger() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	if timeout.isExpired {
		return ErrAlreadyTriggered
	}
	timeout.isExpired = true
	close(timeout.triggeredChannel)
	return nil
}

func (timeout *Timeout) Refresh(timeoutNs int64) error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	if timeout.isExpired {
		return ErrAlreadyTriggered
	}
	timeout.interactionChannel <- timeoutNs
	return nil
}

func (timeout *Timeout) Cancel() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	if timeout.isExpired {
		return ErrAlreadyTriggered
	}
	if !timeout.cancellable {
		return errors.New("timeout cannot be cancelled")
	}
	timeout.isExpired = true
	timeout.interactionChannel <- 0
	return nil
}
