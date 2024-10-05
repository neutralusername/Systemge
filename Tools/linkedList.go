package Tools

type LinkedList struct {
	head *linkedListItem
	tail *linkedListItem
}

type linkedListItem struct {
	next     *linkedListItem
	prev     *linkedListItem
	priority uint32
	item     any
}
