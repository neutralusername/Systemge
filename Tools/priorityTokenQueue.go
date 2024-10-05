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
	token              string
	value              any
	priority           uint32
	deadline           uint64
	isRetrievedChannel chan struct{}
	index              int
}

func NewPriorityTokenQueueItem(token string, value any, priority uint32, deadlineMs uint64) *priorityTokenQueueItem {
	return &priorityTokenQueueItem{
		token:              token,
		value:              value,
		priority:           priority,
		deadline:           deadlineMs,
		isRetrievedChannel: make(chan struct{}),
	}
}

func NewPriorityTokenQueue(priorityQueue PriorityQueue) *PriorityTokenQueue {
	if priorityQueue == nil {
		priorityQueue = make(PriorityQueue, 0)
	}
	queue := &PriorityTokenQueue{
		items:         make(map[string]*priorityTokenQueueItem),
		priorityQueue: priorityQueue,
	}
	heap.Init(&queue.priorityQueue)
	return queue
}

// token may be empty string
func (queue *PriorityTokenQueue) AddItem(token string, value any, priority uint32, deadlineMs uint64) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	item := NewPriorityTokenQueueItem(token, value, priority, deadlineMs)
	if token != "" {
		if queue.items[token] != nil {
			return errors.New("token already exists")
		}
		queue.items[item.token] = item
	}
	heap.Push(&queue.priorityQueue, item)

	if deadlineMs > 0 {
		go func() {
			for {
				select {
				case <-time.After(time.Duration(deadlineMs) * time.Millisecond):
					queue.mutex.Lock()
					select {
					case <-item.isRetrievedChannel:
						queue.mutex.Unlock()
						return
					default:
						queue.removeItem(item)
						queue.mutex.Unlock()
					}
				case <-item.isRetrievedChannel:
				}
			}
		}()
	}
	return nil
}
func (queue *PriorityTokenQueue) removeItem(item *priorityTokenQueueItem) {
	close(item.isRetrievedChannel)
	delete(queue.items, item.token)
	heap.Remove(&queue.priorityQueue, item.index)
}
func (queue *PriorityTokenQueue) GetItemByToken(token string) (any, error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	item := queue.items[token]
	if item == nil {
		return nil, errors.New("token not found")
	}
	queue.removeItem(item)
	return item.value, nil
}

func (queue *PriorityTokenQueue) GetNextItem() (any, error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if len(queue.priorityQueue) == 0 {
		return nil, errors.New("queue is empty")
	}
	item := heap.Pop(&queue.priorityQueue).(*priorityTokenQueueItem)
	close(item.isRetrievedChannel)
	delete(queue.items, item.token)
	return item.value, nil
}

type PriorityQueue []*priorityTokenQueueItem

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
