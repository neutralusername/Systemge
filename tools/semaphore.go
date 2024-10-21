package tools

import "errors"

type Semaphore[T any] struct {
	channel chan T
}

func (semaphore *Semaphore[T]) AvailableAcquires() int {
	return len(semaphore.channel)
}

func NewSemaphore[T any](maxAvailableAcquires int, initialItems []T) (*Semaphore[T], error) {
	if maxAvailableAcquires <= 0 {
		return nil, errors.New("maxAvailableAcquires must be greater than 0")
	}
	if len(initialItems) > maxAvailableAcquires {
		return nil, errors.New("initialItems must be less than or equal to maxAvailableAcquires")
	}
	channel := make(chan T, maxAvailableAcquires)
	for _, item := range initialItems {
		channel <- item
	}
	return &Semaphore[T]{
		channel: channel,
	}, nil
}

// receiving equals Wait.
// sending equals Signal.
func (semaphore *Semaphore[T]) GetChannel() chan T {
	return semaphore.channel
}

// Wait blocks until an item is available and returns it.
func (semaphore *Semaphore[T]) Wait() T {
	return <-semaphore.channel
}

// TryWait returns an item if available.
// If no item is available, it will return an error.
func (semaphore *Semaphore[T]) TryWait() (T, error) {
	select {
	case item := <-semaphore.channel:
		return item, nil
	default:
		var t T
		return t, errors.New("no item available")
	}
}

// Signal blocks until an item can be added to the semaphore.
func (semaphore *Semaphore[T]) Signal(item T) {
	semaphore.channel <- item
}

// Signal adds an item to the semaphore.
// If no space is available, it will return an error.
func (semaphore *Semaphore[T]) TrySignal(item T) error {
	select {
	case semaphore.channel <- item:
		return nil
	default:
		return errors.New("no space available")
	}
}
