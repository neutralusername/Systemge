package Tools

import (
	"container/heap"
	"errors"
	"sync"
)

type PriorityQueue[T any] struct {
	mutex         sync.Mutex
	priorityQueue priorityQueue[T]
}

func NewPriorityQueue[T any]() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		priorityQueue: make(priorityQueue[T], 0),
	}
}

func (priorityQueue *PriorityQueue[T]) Len() int {
	return len(priorityQueue.priorityQueue)
}

func (priorityQueue *PriorityQueue[T]) Push(element T, priority uint32) {
	priorityQueue.mutex.Lock()
	defer priorityQueue.mutex.Unlock()
	heap.Push(&priorityQueue.priorityQueue, &priorityQueueElement[T]{
		value:    element,
		priority: priority,
	})
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
