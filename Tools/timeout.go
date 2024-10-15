package Tools

import (
	"errors"
	"sync"
	"time"
)

var ErrAlreadyExpired = errors.New("timeout already expired")

type Timeout struct {
	onTrigger func()

	timeoutNs int64

	triggerTimestamp time.Time

	cancellable bool

	interactionChannel chan int64
	isExpiredChannel   chan struct{}

	mutex sync.Mutex
}

// timeoutNs 0 == must be triggered manually
func NewTimeout(timeoutNs int64, onTrigger func(), cancellable bool) *Timeout {
	timeout := &Timeout{
		timeoutNs:          timeoutNs,
		onTrigger:          onTrigger,
		interactionChannel: make(chan int64),
		cancellable:        cancellable,
		isExpiredChannel:   make(chan struct{}),
	}
	go timeout.handleTrigger()
	return timeout
}

func (timeout *Timeout) handleTrigger() {
	for {
		var timeoutChannel <-chan time.Time

		if timeout.timeoutNs > 0 {
			timeout.triggerTimestamp = time.Now().Add(time.Duration(timeout.timeoutNs))
			timeoutChannel = time.After(timeout.triggerTimestamp.Sub(time.Now()))
		} else {
			timeout.triggerTimestamp = time.Time{}
		}

		select {
		case newTimeoutNs, ok := <-timeout.interactionChannel:
			if ok {
				timeout.timeoutNs = newTimeoutNs
				continue
			} else {
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
	return timeout.triggerTimestamp
}

func (timeout *Timeout) IsExpired() bool {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	select {
	case <-timeout.isExpiredChannel:
		return true
	default:
		return false
	}
}

func (timeout *Timeout) IsCancellable() bool {
	return timeout.cancellable
}

func (timeout *Timeout) SetOnTrigger(onTrigger func()) error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	select {
	case <-timeout.isExpiredChannel:
		return ErrAlreadyExpired
	default:
	}

	timeout.onTrigger = onTrigger
	return nil
}

// channel will be closed once the timeout is either triggered or cancelled
func (timeout *Timeout) GetIsExpiredChannel() <-chan struct{} {
	return timeout.isExpiredChannel
}

func (timeout *Timeout) Refresh(timeoutNs int64) error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	select {
	case <-timeout.isExpiredChannel:
		return ErrAlreadyExpired
	default:
	}

	timeout.interactionChannel <- timeoutNs
	return nil
}

func (timeout *Timeout) Trigger() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	select {
	case <-timeout.isExpiredChannel:
		return ErrAlreadyExpired
	default:
	}

	close(timeout.interactionChannel)
	close(timeout.isExpiredChannel)
	timeout.onTrigger()
	return nil
}

func (timeout *Timeout) Cancel() error {
	timeout.mutex.Lock()
	defer timeout.mutex.Unlock()

	if !timeout.cancellable {
		return errors.New("timeout cannot be cancelled")
	}

	select {
	case <-timeout.isExpiredChannel:
		return ErrAlreadyExpired
	default:
	}

	close(timeout.interactionChannel)
	close(timeout.isExpiredChannel)
	return nil
}
