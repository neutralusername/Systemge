package Tools

// dynamic buffer:
// items have an optional deadline, after which they are removed from the queue

// AddItem (any, priority, deadlineMs) (token, error)

// GetItem (order/priority)
// GetItemByToken
//

type DynamicBuffer struct {
	items    map[string]*dyamicBufferItem
	order    []*dyamicBufferItem
	itemChan chan *dyamicBufferItem
}

type dyamicBufferItem struct {
	token    string
	item     any
	priority uint32
	deadline uint64
}

func NewDynamicBuffer(capacity uint32) *DynamicBuffer {
	buffer := &DynamicBuffer{
		items:    make(map[string]*dyamicBufferItem, capacity),
		order:    make([]*dyamicBufferItem, 0, capacity),
		itemChan: make(chan *dyamicBufferItem, capacity),
	}
	/* go func() {
		for item := range buffer.itemChan {
			buffer.items[item.token] = item
			for i := len(buffer.order) - 1; i >= 0; i-- {
				if buffer.order[i].priority >= item.priority {
					buffer.order = append(buffer.order, nil)
					copy(buffer.order[i+1:], buffer.order[i:])
					buffer.order[i] = item
					break
				}
			}
		}
	}() */
	return buffer
}

func (buffer *DynamicBuffer) AddItem(token string, item any, priority uint32, deadlineMs uint64) (string, error) {
	buffer.itemChan <- &dyamicBufferItem{
		token:    token,
		item:     item,
		priority: priority,
		deadline: deadlineMs,
	}
	return "", nil
}
