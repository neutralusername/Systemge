package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Status"
)

type QueueConsumerFunc[T any] func(T)

type QueueConsumer[T any] struct {
	status      int
	statusMutex sync.Mutex

	queue   *PriorityTokenQueue[T]
	handler QueueConsumerFunc[T]
}

func NewQueueConsumer[T any](queue *PriorityTokenQueue[T], handler QueueConsumerFunc[T]) *QueueConsumer[T] {
	return &QueueConsumer[T]{
		queue:   queue,
		handler: handler,
	}
}

func (c *QueueConsumer[T]) Start() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()

	if c.status == Status.Stoped {
		return
	}

	c.status = Status.Started
}

func (c *QueueConsumer[T]) Stop() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()

	if c.status == Status.Stoped {
		return
	}

	c.status = Status.Stoped
}

func (c *QueueConsumer[T]) consumeRoutine() {
	for {
		c.handler(c.queue.PopBlocking())
	}
}
