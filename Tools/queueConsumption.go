package Tools

type IQueueConsumption[T any] interface {
	Pop() T
	PopBlocking() T
	PopChannel() <-chan T
	Len() int
}
