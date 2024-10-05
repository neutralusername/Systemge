package Tools

// dynamic buffer:
// items have an optional deadline, after which they are removed from the queue

// AddItem (any, priority, deadlineMs) (token, error)

// GetItem (order/priority)
// GetItemByToken
//

type DynamicBuffer struct {
	items    map[string]*dyamicBufferItem
	head     *dyamicBufferItem // next item to be retrieved
	tail     *dyamicBufferItem // lowest priority item/last item added
	itemChan chan *dyamicBufferItem
}

type dyamicBufferItem struct {
	next     *dyamicBufferItem
	prev     *dyamicBufferItem
	token    string
	item     any
	priority uint32
	deadline uint64
}

func NewDynamicBuffer(capacity uint32) *DynamicBuffer {
	buffer := &DynamicBuffer{
		items:    make(map[string]*dyamicBufferItem, capacity),
		itemChan: make(chan *dyamicBufferItem, capacity),
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

func (buffer *DynamicBuffer) AddItem(token string, item any, priority uint32, deadlineMs uint64) error {
	buffer.itemChan <- &dyamicBufferItem{
		token:    token,
		item:     item,
		priority: priority,
		deadline: deadlineMs,
	}
	return nil
}
