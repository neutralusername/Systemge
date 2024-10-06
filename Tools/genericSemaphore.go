package Tools

import "errors"

type GenericSemaphore[T comparable] struct {
	channel chan T
}

func (semaphore *GenericSemaphore[T]) AvailableAcquires() uint32 {
	return uint32(cap(semaphore.channel)) - uint32(len(semaphore.channel))
}

func NewGenericSemaphore[T comparable](maxAvailableAcquires uint32, initialItems []T) (*GenericSemaphore[T], error) {
	if maxAvailableAcquires == 0 {
		return nil, errors.New("maxAvailableAcquires must be greater than 0")
	}
	if len(initialItems) > int(maxAvailableAcquires) {
		return nil, errors.New("initialItems must be less than or equal to maxAvailableAcquires")
	}
	channel := make(chan T, maxAvailableAcquires)
	for _, item := range initialItems {
		channel <- item
	}
	return &GenericSemaphore[T]{
		channel: channel,
	}, nil
}

// receiving equals to AcquireBlocking and sending equals to ReleaseBlocking
func (semaphore *GenericSemaphore[T]) GetChannel() chan T {
	return semaphore.channel
}

func (semaphore *GenericSemaphore[T]) Wait() T {
	return <-semaphore.channel
}

func (semaphore *GenericSemaphore[T]) TryWait() (T, error) {
	select {
	case item := <-semaphore.channel:
		return item, nil
	default:
		var t T
		return t, errors.New("no item available")
	}
}

func (semaphore *GenericSemaphore[T]) TrySignal(item T) error {
	select {
	case semaphore.channel <- item:
		return nil
	default:
		return errors.New("no space available")
	}
}

func (semaphore *GenericSemaphore[T]) Signal(item T) {
	semaphore.channel <- item
}
