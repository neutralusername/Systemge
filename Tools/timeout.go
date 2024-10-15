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

	triggerTimestamp time.Time

	cancellable bool

	interactionChannel chan int64
	expiredChannel     chan struct{}

	mutex sync.Mutex
}

// timeoutNs 0 == must be triggered manually
func NewTimeout(timeoutNs int64, onTrigger func(), cancellable bool) *Timeout {
	timeout := &Timeout{
		timeoutNs:          timeoutNs,
		onTrigger:          onTrigger,
		interactionChannel: make(chan int64),
		cancellable:        cancellable,
		expiredChannel:     make(chan struct{}),
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
			timeout.triggerTimestamp = time.Unix(0, triggerTimestamp)
		} else {
			timeout.triggerTimestamp = time.Time{}
		}

		select {
		case newTimeoutNs := <-timeout.interactionChannel:
			if newTimeoutNs > 0 {
				timeout.timeoutNs = newTimeoutNs
				continue
			} else {
				close(timeout.expiredChannel)
				return
			}

		case <-timeoutChannel:
			timeout.Trigger()
			return
		}
	}
}

func (timeout *Timeout) GetTimeoutNs() int64 {
	return timeout.timeoutNs
}

func (timeout *Timeout) TriggerTimestamp() time.Time {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	return timeout.triggerTimestamp
}

func (timeout *Timeout) IsExpired() bool {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	select {
	case <-timeout.expiredChannel:
		return true
	default:
		return false
	}
}

func (timeout *Timeout) GetExpiredChannel() <-chan struct{} {
	return timeout.expiredChannel
}

func (timeout *Timeout) Trigger() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	select {
	case <-timeout.expiredChannel:
		return ErrAlreadyTriggered
	default:
	}

	timeout.onTrigger()
	close(timeout.expiredChannel)

	return nil
}

func (timeout *Timeout) Refresh(timeoutNs int64) error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	select {
	case <-timeout.expiredChannel:
		return ErrAlreadyTriggered
	default:
	}

	timeout.interactionChannel <- timeoutNs
	return nil
}

func (timeout *Timeout) Cancel() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	if !timeout.cancellable {
		return errors.New("timeout cannot be cancelled")
	}

	select {
	case <-timeout.expiredChannel:
		return ErrAlreadyTriggered
	default:
	}

	timeout.interactionChannel <- 0
	return nil
}
