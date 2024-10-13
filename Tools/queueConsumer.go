package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Status"
)

type QueueConsumerFunc[T any] func(T)

type QueueConsumer[T any] struct {
	status      int
	statusMutex sync.Mutex
	waitgroup   sync.WaitGroup
	stopChannel chan struct{}

	queue   IQueueConsumption[T]
	handler QueueConsumerFunc[T]
}

func NewQueueConsumer[T any](queue IQueueConsumption[T], handler QueueConsumerFunc[T]) *QueueConsumer[T] {
	return &QueueConsumer[T]{
		queue:   queue,
		handler: handler,
	}
}

func (c *QueueConsumer[T]) Start() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()

	if c.status == Status.Stopped {
		return
	}

	c.stopChannel = make(chan struct{})
	c.status = Status.Started

	c.waitgroup.Add(1)
	go c.consumeRoutine()
}

func (c *QueueConsumer[T]) Stop() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()

	if c.status == Status.Stopped {
		return
	}

	close(c.stopChannel)
	c.waitgroup.Wait()

	c.status = Status.Stopped
}

func (c *QueueConsumer[T]) consumeRoutine() {
	defer c.waitgroup.Done()
	for {
		select {
		case <-c.stopChannel:
			return
		case item := <-c.queue.PopChannel():
			c.handler(item)
		}
	}
}
