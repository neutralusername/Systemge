package Tools

type QueueConsumer[O any] struct {
	queue   PriorityTokenQueue[O]
	handler func(O)
}
