package Tools

import (
	"errors"
	"sync"
)

type PriorityQueue struct {
	items map[string]*priorityQueueLinkedListItem
	head  *priorityQueueLinkedListItem
	tail  *priorityQueueLinkedListItem
	mutex sync.Mutex
}

type priorityQueueLinkedListItem struct {
	next     *priorityQueueLinkedListItem
	prev     *priorityQueueLinkedListItem
	token    string
	item     any
	priority uint32
	deadline uint64
}

func NewDynamicBuffer(capacity uint32) *PriorityQueue {
	buffer := &PriorityQueue{
		items: make(map[string]*priorityQueueLinkedListItem, capacity),
	}
	return buffer
}

func (buffer *PriorityQueue) GetNextItem() (any, error) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	if buffer.head == nil {
		return nil, errors.New("buffer is empty")
	}
	item := buffer.head

	buffer.head = buffer.head.prev
	if buffer.head == nil {
		buffer.tail = nil
	}
	if item.prev != nil {
		item.prev.next = nil
	}
	delete(buffer.items, item.token)
	return item.item, nil
}

func (buffer *PriorityQueue) GetItemByToken(token string) (any, error) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	item := buffer.items[token]
	if item == nil {
		return nil, errors.New("item not found")
	}
	if item.prev != nil {
		item.prev.next = item.next
	}
	if item.next != nil {
		item.next.prev = item.prev
	}
	if buffer.head == item {
		buffer.head = item.prev
	}
	if buffer.tail == item {
		buffer.tail = item.next
	}
	delete(buffer.items, item.token)
	return item.item, nil
}

func (buffer *PriorityQueue) AddItem(token string, item any, priority uint32, deadlineMs uint64) error {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	if buffer.items[token] != nil {
		return errors.New("token already exists")
	}
	linkedListItem := &priorityQueueLinkedListItem{
		token:    token,
		item:     item,
		priority: priority,
		deadline: deadlineMs,
	}
	buffer.items[linkedListItem.token] = linkedListItem
	if buffer.tail == nil {
		buffer.head = linkedListItem
		buffer.tail = linkedListItem
		return nil
	}
	if linkedListItem.priority > buffer.head.priority {
		linkedListItem.prev = buffer.head
		buffer.head.next = linkedListItem
		buffer.head = linkedListItem
		return nil
	}
	current := buffer.tail
	for current.priority < linkedListItem.priority {
		current = current.next
	}
	linkedListItem.next = current
	linkedListItem.prev = current.prev
	current.prev.next = linkedListItem
	current.prev = linkedListItem
	return nil
}
