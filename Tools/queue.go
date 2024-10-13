package Tools

type Queue[T any] interface {
	Push(T)
	Pop() T
	PopBlocking() T
	PopChannel() <-chan T
	Len() int
}
