package Tools

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

type PriorityTokenQueue[T any] struct {
	elements      map[string]*priorityQueueElement[*tokenItem[T]]
	mutex         sync.Mutex
	priorityQueue priorityQueue[*tokenItem[T]]

	maxElements   uint32
	replaceIfFull bool
}

type tokenItem[T any] struct {
	item               T
	token              string
	isRetrievedChannel chan struct{}
}

func NewPriorityTokenQueue[T any](maxElements uint32, replaceIfFull bool) *PriorityTokenQueue[T] {
	queue := &PriorityTokenQueue[T]{
		elements:      make(map[string]*priorityQueueElement[*tokenItem[T]]),
		priorityQueue: make(priorityQueue[*tokenItem[T]], 0),
		maxElements:   maxElements,
		replaceIfFull: replaceIfFull,
	}
	heap.Init(&queue.priorityQueue)
	return queue
}

func newTokenItem[T any](token string, value T) *tokenItem[T] {
	return &tokenItem[T]{
		item:               value,
		token:              token,
		isRetrievedChannel: make(chan struct{}),
	}
}

// Push pushes a new item into the queue based on its priority.
// If the queue is full and replaceIfFull is false, an error is returned.
// The token may be an emptry string, which means that the item cannot be retrieved by token.
// If the token is not empty and already exists, an error is returned.
// If a timeout is set, the item will be removed from the queue after the timeout.
func (queue *PriorityTokenQueue[T]) Push(token string, value T, priority uint32, timeoutMs uint64) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if queue.maxElements > 0 && uint32(len(queue.priorityQueue)) >= queue.maxElements {
		if !queue.replaceIfFull {
			return errors.New("priority queue is full")
		}
		heap.Pop(&queue.priorityQueue)
	}

	element := &priorityQueueElement[*tokenItem[T]]{
		value:    newTokenItem(token, value),
		priority: priority,
	}
	if token != "" {
		if queue.elements[token] != nil {
			return errors.New("token already exists")
		}
		queue.elements[token] = element
	}
	heap.Push(&queue.priorityQueue, element)

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

// Pop returns the next item from the queue.
// If the queue is empty, an error is returned.
func (queue *PriorityTokenQueue[T]) Pop() (T, error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if len(queue.priorityQueue) == 0 {
		var nilValue T
		return nilValue, errors.New("queue is empty")
	}
	element := heap.Pop(&queue.priorityQueue).(*priorityQueueElement[*tokenItem[T]])
	close(element.value.isRetrievedChannel)
	delete(queue.elements, element.value.token)
	return element.value.item, nil
}

// PopToken returns the item with the given token from the queue.
// If the token does not exist, an error is returned.
func (queue *PriorityTokenQueue[T]) PopToken(token string) (T, error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	element, ok := queue.elements[token]
	if !ok {
		var nilValue T
		return nilValue, errors.New("token not found")
	}
	queue.remove(element)
	return element.value.item, nil
}

// Len returns the number of items in the queue.
func (queue *PriorityTokenQueue[T]) Len() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	return len(queue.priorityQueue)
}

func (queue *PriorityTokenQueue[T]) remove(element *priorityQueueElement[*tokenItem[T]]) {
	close(element.value.isRetrievedChannel)
	delete(queue.elements, element.value.token)
	heap.Remove(&queue.priorityQueue, element.index)
}
