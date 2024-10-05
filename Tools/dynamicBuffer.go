package Tools

import "errors"

// dynamic buffer:
// items have an optional deadline, after which they are removed from the queue

// AddItem (any, priority, deadlineMs) (token, error)

// GetItem (order/priority)
// GetItemByToken
//

type DynamicBuffer struct {
	items    map[string]*dynamicBufferItem
	head     *dynamicBufferItem // next item to be retrieved
	tail     *dynamicBufferItem // lowest priority item/last item added
	itemChan chan *dynamicBufferItem
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
		items:    make(map[string]*dynamicBufferItem, capacity),
		itemChan: make(chan *dynamicBufferItem, capacity),
	}
	go func() {
		for item := range buffer.itemChan {
			if buffer.items[item.token] != nil {
				//
				continue
			}
			buffer.items[item.token] = item
			if buffer.tail == nil {
				buffer.head = item
				buffer.tail = item
				continue
			}
			if item.priority > buffer.head.priority {
				item.prev = buffer.head
				buffer.head.next = item
				buffer.head = item
				continue
			}
			current := buffer.tail
			for {
				if current.priority >= item.priority {
					item.next = current
					item.prev = current.prev
					current.prev.next = item
					current.prev = item
					break
				}
				current = current.next
			}
		}
	}()
	return buffer
}

func (buffer *DynamicBuffer) GetNextItem() (any, error) {
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
	buffer.itemChan <- &dynamicBufferItem{
		token:    token,
		item:     item,
		priority: priority,
		deadline: deadlineMs,
	}
	return nil
}
