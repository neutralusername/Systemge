package Tools

import (
	"errors"
	"sync"
)

type OrderedMap[K comparable, V any] struct {
	head   *orderedMapNode[K, V]
	tail   *orderedMapNode[K, V]
	values map[K]*orderedMapNode[K, V]
	mutex  sync.Mutex
}

type orderedMapNode[K comparable, V any] struct {
	key   K
	value V
	next  *orderedMapNode[K, V]
	prev  *orderedMapNode[K, V]
}

func NewOrderedMap[K comparable, V any](initialItems map[K]V) *OrderedMap[K, V] {
	orderedMap := &OrderedMap[K, V]{
		values: make(map[K]*orderedMapNode[K, V]),
	}
	for key, value := range initialItems {
		orderedMap.Push(key, value)
	}
	return orderedMap
}

func newOrderedMapNode[K comparable, V any](key K, value V) *orderedMapNode[K, V] {
	return &orderedMapNode[K, V]{
		key:   key,
		value: value,
	}
}

func (orderedMap *OrderedMap[K, V]) Push(key K, value V) error {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

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

func (orderedMap *OrderedMap[K, V]) Pop() (K, V, error) {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

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

func (orderedMap *OrderedMap[K, V]) Get(key K) (V, error) {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

	if node, ok := orderedMap.values[key]; ok {
		return node.value, nil
	}
	var nilValue V
	return nilValue, errors.New("key not found")
}

func (orderedMap *OrderedMap[K, V]) Remove(key K) error {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

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

// returns the keys in the order they were pushed
func (orderedMap *OrderedMap[K, V]) GetKeys() []K {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

	keys := make([]K, 0, len(orderedMap.values))
	currentNode := orderedMap.head
	for currentNode != nil {
		keys = append(keys, currentNode.key)
		currentNode = currentNode.next
	}
	return keys
}

// returns the values in the order they were pushed
func (orderedMap *OrderedMap[K, V]) GetValues() []V {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

	values := make([]V, 0, len(orderedMap.values))
	currentNode := orderedMap.head
	for currentNode != nil {
		values = append(values, currentNode.value)
		currentNode = currentNode.next
	}
	return values
}
