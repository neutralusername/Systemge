package Tools

// queuemanger:
// items have an optional deadline, after which they are removed from the queue

// Add (any, priority, deadlineMs) (token, error)

// GetItem (order/priority)
// GetItemFIFO
// GetItemLIFO
// GetItemByToken
//

type QueueManager struct {
	channel chan any
}

func NewReceptionManager(capacity uint32) *QueueManager {
	return &QueueManager{
		channel: make(chan any, capacity),
	}
}

func (manager *QueueManager) Add(item any) (string, error) {

}
