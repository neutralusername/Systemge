package tools

import (
	"github.com/neutralusername/Systemge/Config"
)

type QueueConsumerFunc[T any] func(T)

func NewQueueConsumer[T any](config *Config.Routine, queue IQueueConsumer[T], handler QueueConsumerFunc[T]) *Routine {
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
	return routine
}
