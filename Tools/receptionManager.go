package Tools

// dynamic buffer:
// items have an optional deadline, after which they are removed from the queue

// Add (any, priority, deadlineMs) (token, error)

// GetItem (order/priority)
// GetItemFIFO
// GetItemLIFO
// GetItemByToken
//

type DynamicBuffer struct {
	channel chan any
}

func NewDynamicBuffer(capacity uint32) *DynamicBuffer {
	return &DynamicBuffer{
		channel: make(chan any, capacity),
	}
}

func (buffer *DynamicBuffer) Add(item any, priority uint32, deadlineMs uint64) (string, error) {
}
