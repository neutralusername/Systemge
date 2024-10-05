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
	priorityQueue priorityQueue
}

type priorityTokenQueueItem struct {
	token     string
	value     any
	priority  uint32
	deadline  uint64
	retrieved chan struct{}
	index     int
}

type priorityQueue []*priorityTokenQueueItem

func NewPriorityTokenQueue() *PriorityTokenQueue {
	queue := &PriorityTokenQueue{
		items: make(map[string]*priorityTokenQueueItem),
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

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*priorityTokenQueueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

/*
func (pq *priorityQueue) update(item *priorityTokenQueueItem, value string, priority uint32) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}


// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in priority order.
func main() {
	// Some items and their priorities.
	items := map[string]int{
		"banana": 3, "apple": 2, "pear": 4,
	}

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, len(items))
	i := 0
	for value, priority := range items {
		pq[i] = &Item{
			value:    value,
			priority: priority,
			index:    i,
		}
		i++
	}
	heap.Init(&pq)

	// Insert a new item and then modify its priority.
	item := &Item{
		value:    "orange",
		priority: 1,
	}
	heap.Push(&pq, item)
	pq.update(item, item.value, 5)

	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		fmt.Printf("%.2d:%s ", item.priority, item.value)
	}
}
*/
