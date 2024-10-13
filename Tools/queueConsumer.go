package Tools

type QueueConsumer[O any] struct {
	queue priorityQueue[O]
	handler
}
