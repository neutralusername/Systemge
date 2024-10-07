package Tools

import (
	"container/heap"
	"errors"
	"sync"
)

type PriorityQueue[T any] struct {
	mutex         sync.Mutex
	priorityQueue priorityQueue[T]
	maxItems      uint32
	replaceOnFull bool
}

func NewPriorityQueue[T any](maxItems uint32, replaceOnFull bool) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		priorityQueue: make(priorityQueue[T], 0),
		maxItems:      maxItems,
		replaceOnFull: replaceOnFull,
	}
}

func (priorityQueue *PriorityQueue[T]) Len() int {
	return len(priorityQueue.priorityQueue)
}

func (priorityQueue *PriorityQueue[T]) Push(element T, priority uint32) error {
	priorityQueue.mutex.Lock()
	defer priorityQueue.mutex.Unlock()
	if priorityQueue.maxItems > 0 && uint32(len(priorityQueue.priorityQueue)) >= priorityQueue.maxItems {
		if !priorityQueue.replaceOnFull {
			return errors.New("priority queue is full")
		}
		heap.Pop(&priorityQueue.priorityQueue)
	}
	heap.Push(&priorityQueue.priorityQueue, &priorityQueueElement[T]{
		value:    element,
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
	element := heap.Pop(&priorityQueue.priorityQueue).(*priorityQueueElement[T])
	return element.value, nil
}
