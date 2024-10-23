package tools

import (
	"github.com/neutralusername/systemge/configs"
)

type IQueueConsumer[T any] interface {
	Pop() (T, error)
	PopBlocking() T
	PopChannel() <-chan T
	Len() int
}

type QueueConsumerFunc[T any] func(T)

func NewQueueConsumer[T any](config *configs.Routine, queue IQueueConsumer[T], handler QueueConsumerFunc[T]) (*Routine, error) {
	return NewRoutine(
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
}
