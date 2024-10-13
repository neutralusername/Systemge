package Tools

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Config"
)

type SimpleQueue[T any] struct {
	config  *Config.Queue
	queue   []T
	waiting []chan T
	mutex   sync.Mutex
}

func NewSimpleQueue[T any](config *Config.Queue) *SimpleQueue[T] {
	return &SimpleQueue[T]{
		config: config,
		queue:  make([]T, 0),
	}
}

func (queue *SimpleQueue[T]) Push(value T) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	for _, waiting := range queue.waiting {
		waiting <- value
		close(waiting)
		queue.waiting = queue.waiting[1:]
		return nil
	}

	if queue.config.MaxElements > 0 && uint32(len(queue.queue)) >= queue.config.MaxElements {
		if !queue.config.ReplaceIfFull {
			return errors.New("priority queue is full")
		}
		queue.queue = queue.queue[1:]
	}

	queue.queue = append(queue.queue, value)

	return nil
}

func (queue *SimpleQueue[T]) Pop() (T, error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if len(queue.queue) == 0 {
		var nilValue T
		return nilValue, errors.New("queue is empty")
	}
	item := queue.queue[0]
	queue.queue = queue.queue[1:]
	return item, nil
}

func (queue *SimpleQueue[T]) PopBlocking() T {
	return <-queue.PopChannel()
}

func (queue *SimpleQueue[T]) PopChannel() <-chan T {
	c := make(chan T)
	go func() {

		queue.mutex.Lock()
		if len(queue.queue) == 0 {
			queue.waiting = append(queue.waiting, c)
			queue.mutex.Unlock()
			return
		}

		item := queue.queue[0]
		queue.queue = queue.queue[1:]
		queue.mutex.Unlock()

		c <- item
		close(c)
	}()
	return c
}

func (queue *SimpleQueue[T]) Len() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	return len(queue.queue)
}
