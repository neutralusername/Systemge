package Tools

import (
	"github.com/neutralusername/Systemge/Config"
)

type QueueConsumerFunc[T any] func(T)

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
