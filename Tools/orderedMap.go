package Tools

import "errors"

// O(1) delete
type OrderedMap[K comparable, V any] struct {
	head   *orderedMapNode[K, V]
	tail   *orderedMapNode[K, V]
	values map[K]*orderedMapNode[K, V]
}

type orderedMapNode[K comparable, V any] struct {
	key   K
	value V
	next  *orderedMapNode[K, V]
	prev  *orderedMapNode[K, V]
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		values: make(map[K]*orderedMapNode[K, V]),
	}
}

func newOrderedMapNode[K comparable, V any](key K, value V) *orderedMapNode[K, V] {
	return &orderedMapNode[K, V]{
		key:   key,
		value: value,
	}
}

func (fifoMap *OrderedMap[K, V]) Add(key K, value V) error {
	if _, ok := fifoMap.values[key]; ok {
		return errors.New("key already exists")
	}

	node := newOrderedMapNode(key, value)
	fifoMap.values[key] = node
	if fifoMap.head == nil {
		fifoMap.head = node
		fifoMap.tail = node
		return nil
	}

	fifoMap.tail.next = node
	node.prev = fifoMap.tail
	fifoMap.tail = node
	return nil
}

func (fifoMap *OrderedMap[K, V]) Get(key K) (V, error) {
	if node, ok := fifoMap.values[key]; ok {
		return node.value, nil
	}
	var nilValue V
	return nilValue, errors.New("key not found")
}

func (fifoMap *OrderedMap[K, V]) Delete(key K) error {
	if node, ok := fifoMap.values[key]; ok {
		delete(fifoMap.values, key)
		if node.prev != nil {
			node.prev.next = node.next
		} else {
			fifoMap.head = node.next
		}

		if node.next != nil {
			node.next.prev = node.prev
		} else {
			fifoMap.tail = node.prev
		}
		return nil
	}
	return errors.New("key not found")
}

func (fifoMap *OrderedMap[K, V]) Pop() (K, V, error) {
	if fifoMap.head == nil {
		var nilKey K
		var nilValue V
		return nilKey, nilValue, errors.New("empty map")
	}
	node := fifoMap.head
	delete(fifoMap.values, node.key)
	fifoMap.head = node.next
	if node.next != nil {
		node.next.prev = nil
	} else {
		fifoMap.tail = nil
	}
	return node.key, node.value, nil
}

func (fifoMap *OrderedMap[K, V]) GetKeys() []K {
	keys := make([]K, 0, len(fifoMap.values))
	currentNode := fifoMap.tail
	for currentNode != nil {
		keys = append(keys, currentNode.key)
		currentNode = currentNode.prev
	}
	return keys
}

func (fifoMap *OrderedMap[K, V]) GetValues() []V {
	values := make([]V, 0, len(fifoMap.values))
	currentNode := fifoMap.tail
	for currentNode != nil {
		values = append(values, currentNode.value)
		currentNode = currentNode.prev
	}
	return values
}
