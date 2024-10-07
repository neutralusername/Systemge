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
}

type tokenItem[T any] struct {
	item               T
	token              string
	isRetrievedChannel chan struct{}
}

func NewPriorityTokenQueue[T any]() *PriorityTokenQueue[T] {
	queue := &PriorityTokenQueue[T]{
		elements:      make(map[string]*priorityQueueElement[*tokenItem[T]]),
		priorityQueue: make(priorityQueue[*tokenItem[T]], 0),
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

// token may be empty string
func (queue *PriorityTokenQueue[T]) Add(token string, value T, priority uint32, timeoutMs uint64) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

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

func (queue *PriorityTokenQueue[T]) remove(element *priorityQueueElement[*tokenItem[T]]) {
	close(element.value.isRetrievedChannel)
	delete(queue.elements, element.value.token)
	heap.Remove(&queue.priorityQueue, element.index)
}

func (queue *PriorityTokenQueue[T]) RetrieveByToken(token string) (T, error) {
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

func (queue *PriorityTokenQueue[T]) RetrieveNext() (T, error) {
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
