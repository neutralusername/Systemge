package Tools

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

type PriorityTokenQueue struct {
	items map[string]*priorityTokenQueueItem
	mutex sync.Mutex
	queue []*priorityTokenQueueItem
}

type priorityTokenQueueItem struct {
	token     string
	value     any
	priority  uint32
	deadline  uint64
	retrieved chan struct{}
	index     int
}

func (queue *PriorityTokenQueue) Len() int {
	return len(queue.queue)
}

func (queue *PriorityTokenQueue) Less(i, j int) bool {
	return queue.queue[i].priority > queue.queue[j].priority
}

func (queue *PriorityTokenQueue) Swap(i, j int) {
	queue.queue[i], queue.queue[j] = queue.queue[j], queue.queue[i]
	queue.queue[i].index = i
	queue.queue[j].index = j
}

func (queue *PriorityTokenQueue) Push(x any) {
	n := len(queue.queue)
	item := x.(*priorityTokenQueueItem)
	item.index = n
	queue.queue = append(queue.queue, item)
}

func (queue *PriorityTokenQueue) Pop() any {
	old := queue.queue
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	queue.queue = old[0 : n-1]
	return item
}

func (queue *PriorityTokenQueue) AddItem(token string, value any, priority uint32, deadlineMs uint64) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if queue.items[token] != nil {
		return errors.New("token already exists")
	}
	item := &priorityTokenQueueItem{
		token:     token,
		value:     value,
		priority:  priority,
		deadline:  deadlineMs,
		retrieved: make(chan struct{}),
	}
	queue.items[item.token] = item
	heap.Push(queue, item)

	if deadlineMs > 0 {
		go func() {
			for {
				select {
				case <-time.After(time.Duration(deadlineMs) * time.Millisecond):
					queue.RetrieveItem(token)
					return
				case <-item.retrieved:
					return
				}
			}
		}()
	}
	return nil
}

func (queue *PriorityTokenQueue) RetrieveItem(token string) (any, error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	item := queue.items[token]
	if item == nil {
		return nil, errors.New("token not found")
	}
	close(item.retrieved)
	delete(queue.items, token)
	heap.Remove(queue, item.index)
	return item.value, nil
}

func (queue *PriorityTokenQueue) RetrieveNextItem() (any, error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if len(queue.queue) == 0 {
		return nil, errors.New("queue is empty")
	}
	item := heap.Pop(queue).(*priorityTokenQueueItem)
	close(item.retrieved)
	delete(queue.items, item.token)
	return item.value, nil
}
