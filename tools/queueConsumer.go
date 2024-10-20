package tools

import (
	"github.com/neutralusername/systemge/configs"
)

type QueueConsumerFunc[T any] func(T)

func NewQueueConsumer[T any](config *configs.Routine, queue IQueueConsumer[T], handler QueueConsumerFunc[T]) *Routine {
	routine := NewRoutine(
		func(stopChannel <-chan struct{}) {
			select {
			case <-stopChannel:
				return
			case item := <-queue.PopChannel():
				handler(item)
			}
		},
		config,
	)
	return routine
}
