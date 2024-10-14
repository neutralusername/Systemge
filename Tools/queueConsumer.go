package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
)

type QueueConsumerFunc[T any] func(T)

type QueueConsumer[T any] struct {
	status      int
	statusMutex sync.Mutex
	waitgroup   sync.WaitGroup
	stopChannel chan struct{}
	routine     *Routine
}

func NewQueueConsumer[T any](config *Config.Routine, queue IQueueConsumption[T], handler QueueConsumerFunc[T]) *Routine {
	routine := NewRoutine(
		func(stopChannel <-chan struct{}) {
			select {
			case <-stopChannel:
				return
			case item := <-queue.PopChannel():
				handler(item)
			}
		},
		config.MaxConcurrentHandlers, config.DelayNs, config.TimeoutNs,
	)
	routine.StartRoutine()

	return routine
}
