package tools

type IQueueConsumer[T any] interface {
	Pop() (T, error)
	PopBlocking() T
	PopChannel() <-chan T
	Len() int
}
