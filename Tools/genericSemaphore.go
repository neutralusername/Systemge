package Tools

import "errors"

type GenericSemaphore[T comparable] struct {
	channel chan T
}

func (semaphore *GenericSemaphore[T]) AvailableAcquires() uint32 {
	return uint32(cap(semaphore.channel)) - uint32(len(semaphore.channel))
}

func NewGenericSemaphore2[T comparable](maxAvailableAcquires uint32) *GenericSemaphore[T] {
	if maxAvailableAcquires <= 0 {
		panic("maxValue must be greater than 0")
	}
	channel := make(chan T, maxAvailableAcquires)
	return &GenericSemaphore[T]{
		channel: channel,
	}
}

// receiving equals to AcquireBlocking and sending equals to ReleaseBlocking
func (semaphore *GenericSemaphore[T]) GetChannel() chan T {
	return semaphore.channel
}

func (semaphore *GenericSemaphore[T]) AcquireBlocking() T {
	return <-semaphore.channel
}

func (semaphore *GenericSemaphore[T]) TryAcquire() (T, error) {
	select {
	case item := <-semaphore.channel:
		return item, nil
	default:
		var t T
		return t, errors.New("no item available")
	}
}

func (semaphore *GenericSemaphore[T]) TryRelease(item T) error {
	select {
	case semaphore.channel <- item:
		return nil
	default:
		return errors.New("no space available")
	}
}

func (semaphore *GenericSemaphore[T]) ReleaseBlocking(item T) {
	semaphore.channel <- item
}
