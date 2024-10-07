package Tools

import (
	"container/heap"
	"errors"
	"sync"
)

type PriorityQueue[T any] struct {
	mutex         sync.Mutex
	priorityQueue priorityQueue[T]
	maxElements   uint32
	replaceIfFull bool
}

func NewPriorityQueue[T any](maxElements uint32, replaceIfFull bool) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		priorityQueue: make(priorityQueue[T], 0),
		maxElements:   maxElements,
		replaceIfFull: replaceIfFull,
	}
}

func (priorityQueue *PriorityQueue[T]) Len() int {
	return len(priorityQueue.priorityQueue)
}

func (priorityQueue *PriorityQueue[T]) Push(value T, priority uint32) error {
	priorityQueue.mutex.Lock()
	defer priorityQueue.mutex.Unlock()

	if priorityQueue.maxElements > 0 && uint32(len(priorityQueue.priorityQueue)) >= priorityQueue.maxElements {
		if !priorityQueue.replaceIfFull {
			return errors.New("priority queue is full")
		}
		heap.Pop(&priorityQueue.priorityQueue)
	}
	heap.Push(&priorityQueue.priorityQueue, &priorityQueueElement[T]{
		value:    value,
		priority: priority,
	})
	return nil
}

func (priorityQueue *PriorityQueue[T]) Pop() (T, error) {
	priorityQueue.mutex.Lock()
	defer priorityQueue.mutex.Unlock()

	if len(priorityQueue.priorityQueue) == 0 {
		var nilValue T
		return nilValue, errors.New("priority queue is empty")
	}
	return heap.Pop(&priorityQueue.priorityQueue).(*priorityQueueElement[T]).value, nil
}
