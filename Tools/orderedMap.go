package Tools

import "errors"

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

func (orderedMap *OrderedMap[K, V]) Add(key K, value V) error {
	if _, ok := orderedMap.values[key]; ok {
		return errors.New("key already exists")
	}

	node := newOrderedMapNode(key, value)
	orderedMap.values[key] = node
	if orderedMap.head == nil {
		orderedMap.head = node
		orderedMap.tail = node
		return nil
	}

	orderedMap.tail.next = node
	node.prev = orderedMap.tail
	orderedMap.tail = node
	return nil
}

func (orderedMap *OrderedMap[K, V]) Get(key K) (V, error) {
	if node, ok := orderedMap.values[key]; ok {
		return node.value, nil
	}
	var nilValue V
	return nilValue, errors.New("key not found")
}

func (orderedMap *OrderedMap[K, V]) Remove(key K) error {
	node, ok := orderedMap.values[key]
	if !ok {
		return errors.New("key not found")
	}

	delete(orderedMap.values, key)
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		orderedMap.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		orderedMap.tail = node.prev
	}
	return nil
}

func (orderedMap *OrderedMap[K, V]) Pop() (K, V, error) {
	if orderedMap.head == nil {
		var nilKey K
		var nilValue V
		return nilKey, nilValue, errors.New("empty map")
	}
	node := orderedMap.head
	delete(orderedMap.values, node.key)
	orderedMap.head = node.next
	if node.next != nil {
		node.next.prev = nil
	} else {
		orderedMap.tail = nil
	}
	return node.key, node.value, nil
}

func (orderedMap *OrderedMap[K, V]) GetKeys() []K {
	keys := make([]K, 0, len(orderedMap.values))
	currentNode := orderedMap.tail
	for currentNode != nil {
		keys = append(keys, currentNode.key)
		currentNode = currentNode.prev
	}
	return keys
}

func (orderedMap *OrderedMap[K, V]) GetValues() []V {
	values := make([]V, 0, len(orderedMap.values))
	currentNode := orderedMap.tail
	for currentNode != nil {
		values = append(values, currentNode.value)
		currentNode = currentNode.prev
	}
	return values
}