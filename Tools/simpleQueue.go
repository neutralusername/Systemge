package Tools

import (
	"container/heap"
	"errors"
	"sync"
	"time"

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
	heap.Push(&queue.queue, element)

	if timeoutMs > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
				queue.mutex.Lock()
				defer queue.mutex.Unlock()
				select {
				case <-element.value.isRetrievedChannel:
				default:
					queue.remove(element)
				}
			case <-element.value.isRetrievedChannel:
			}
		}()
	}
	return nil
}
