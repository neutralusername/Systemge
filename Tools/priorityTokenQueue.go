package Tools

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

type PriorityTokenQueue struct {
	items         map[string]*priorityTokenQueueItem
	mutex         sync.Mutex
	priorityQueue PriorityQueue
}

type priorityTokenQueueItem struct {
	token     string
	value     any
	priority  uint32
	deadline  uint64
	retrieved chan struct{}
	index     int
}

func NewPriorityTokenQueue() *PriorityTokenQueue {
	queue := &PriorityTokenQueue{
		items:         make(map[string]*priorityTokenQueueItem),
		priorityQueue: make(PriorityQueue, 0),
	}
	heap.Init(&queue.priorityQueue)
	return queue
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
	heap.Push(&queue.priorityQueue, item)

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
	heap.Remove(&queue.priorityQueue, item.index)
	return item.value, nil
}

func (queue *PriorityTokenQueue) RetrieveNextItem() (any, error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if len(queue.priorityQueue) == 0 {
		return nil, errors.New("queue is empty")
	}
	item := heap.Pop(&queue.priorityQueue).(*priorityTokenQueueItem)
	close(item.retrieved)
	delete(queue.items, item.token)
	return item.value, nil
}

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*priorityTokenQueueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

type PriorityQueue []*priorityTokenQueueItem
