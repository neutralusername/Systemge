package Tools

/* type PriorityTokenQueue struct {
	items map[string]*priorityTokenQueueItem
	head  *priorityTokenQueueItem
	tail  *priorityTokenQueueItem
	mutex sync.Mutex
}

type priorityTokenQueueItem struct {
	next      *priorityTokenQueueItem
	prev      *priorityTokenQueueItem
	token     string
	item      any
	priority  uint32
	deadline  uint64
	retrieved chan struct{}
}

func NewPriorityTokenQueue() *PriorityTokenQueue {
	buffer := &PriorityTokenQueue{
		items: make(map[string]*priorityTokenQueueItem),
	}
	return buffer
}

func (buffer *PriorityTokenQueue) AddItem(token string, item any, priority uint32, deadlineMs uint64) error {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	if buffer.items[token] != nil {
		return errors.New("token already exists")
	}
	linkedListItem := &priorityTokenQueueItem{
		token:     token,
		item:      item,
		priority:  priority,
		deadline:  deadlineMs,
		retrieved: make(chan struct{}),
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

	if deadlineMs > 0 {
		go func() {
			for {
				select {
				case <-time.After(time.Duration(deadlineMs) * time.Millisecond):
					buffer.GetItemByToken(token)
					return
				case <-linkedListItem.retrieved:
					return
				}
			}
		}()
	}
	return nil
}

func (buffer *PriorityTokenQueue) GetNextItem() (any, error) {
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
	close(item.retrieved)
	return item.item, nil
}

func (buffer *PriorityTokenQueue) GetItemByToken(token string) (any, error) {
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
	close(item.retrieved)
	return item.item, nil
}
*/
