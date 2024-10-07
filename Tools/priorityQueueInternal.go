package Tools

type priorityQueueElement[T comparable] struct {
	value    T
	priority uint32
	index    int
}
type priorityQueue[T comparable] []*priorityQueueElement[T]

func (priorityQueue priorityQueue[T]) Len() int {
	return len(priorityQueue)
}

func (priorityQueue priorityQueue[T]) Less(i, j int) bool {
	return priorityQueue[i].priority > priorityQueue[j].priority
}

func (priorityQueue priorityQueue[T]) Swap(i, j int) {
	priorityQueue[i], priorityQueue[j] = priorityQueue[j], priorityQueue[i]
	priorityQueue[i].index = i
	priorityQueue[j].index = j
}

func (priorityQueue *priorityQueue[T]) Push(element any) {
	n := len(*priorityQueue)
	typedElement := element.(*priorityQueueElement[T])
	typedElement.index = n
	*priorityQueue = append(*priorityQueue, typedElement)
}

func (priorityQueue *priorityQueue[T]) Pop() any {
	oldPriorityQueue := *priorityQueue
	elementsCount := len(oldPriorityQueue)
	element := oldPriorityQueue[elementsCount-1]
	oldPriorityQueue[elementsCount-1] = nil
	element.index = -1
	*priorityQueue = oldPriorityQueue[0 : elementsCount-1]
	return element
}
