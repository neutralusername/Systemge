package Tools

type QueueConsumer[T any] struct {
	queue   PriorityTokenQueue[T]
	handler func(T)
}
