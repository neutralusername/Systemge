package Tools

import (
	"errors"
	"sync"
)

// dynamic buffer:
// items have an optional deadline, after which they are removed from the queue

// AddItem (any, priority, deadlineMs) (token, error)

// GetItem (order/priority)
// GetItemByToken
//

type DynamicBuffer struct {
	items map[string]*dynamicBufferItem
	head  *dynamicBufferItem // next item to be retrieved
	tail  *dynamicBufferItem // lowest priority item/last item added
	mutex sync.Mutex
}

type dynamicBufferItem struct {
	next     *dynamicBufferItem
	prev     *dynamicBufferItem
	token    string
	item     any
	priority uint32
	deadline uint64
}

func NewDynamicBuffer(capacity uint32) *DynamicBuffer {
	buffer := &DynamicBuffer{
		items: make(map[string]*dynamicBufferItem, capacity),
	}
	return buffer
}

func (buffer *DynamicBuffer) GetNextItem() (any, error) {
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

func (buffer *DynamicBuffer) GetItemByToken(token string) (any, error) {
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

func (buffer *DynamicBuffer) AddItem(token string, item any, priority uint32, deadlineMs uint64) error {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	if buffer.items[token] != nil {
		return errors.New("token already exists")
	}
	linkedListItem := &dynamicBufferItem{
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
	for {
		if current.priority >= linkedListItem.priority {
			linkedListItem.next = current
			linkedListItem.prev = current.prev
			current.prev.next = linkedListItem
			current.prev = linkedListItem
			return nil
		}
		current = current.next
	}
}
