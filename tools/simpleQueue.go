package tools

import (
	"errors"
	"sync"

	"github.com/neutralusername/systemge/configs"
)

type SimpleFifoQueue[T any] struct {
	config  *configs.Queue
	queue   []T
	waiting []chan T
	mutex   sync.Mutex
}

func NewSimpleQueue[T any](config *configs.Queue) *SimpleFifoQueue[T] {
	return &SimpleFifoQueue[T]{
		config: config,
		queue:  make([]T, 0),
	}
}

func (queue *SimpleFifoQueue[T]) Push(value T) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	for _, waiting := range queue.waiting {
		waiting <- value
		close(waiting)
		queue.waiting = queue.waiting[1:]
		return nil
	}

	if queue.config.MaxElements > 0 && len(queue.queue) >= queue.config.MaxElements {
		if !queue.config.ReplaceIfFull {
			return errors.New("priority queue is full")
		}
		queue.queue = queue.queue[1:]
	}

	queue.queue = append(queue.queue, value)

	return nil
}

func (queue *SimpleFifoQueue[T]) Pop() (T, error) {
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

func (queue *SimpleFifoQueue[T]) PopBlocking() T {
	return <-queue.PopChannel()
}

func (queue *SimpleFifoQueue[T]) PopChannel() <-chan T {
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

func (queue *SimpleFifoQueue[T]) Len() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	return len(queue.queue)
}
