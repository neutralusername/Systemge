package Tools

import (
	"errors"
	"sync"
)

type OrderedMap[K comparable, V any] struct {
	head   *orderedMapNode[K, V]
	tail   *orderedMapNode[K, V]
	values map[K]*orderedMapNode[K, V]
	mutex  sync.RWMutex
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
		orderedMap.Add(key, value)
	}
	return orderedMap
}

func newOrderedMapNode[K comparable, V any](key K, value V) *orderedMapNode[K, V] {
	return &orderedMapNode[K, V]{
		key:   key,
		value: value,
	}
}

func (orderedMap *OrderedMap[K, V]) Add(key K, value V) error {
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

func (orderedMap *OrderedMap[K, V]) RetrieveFIFO() (K, V, error) {
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

func (orderedMap *OrderedMap[K, V]) RetrieveLIFO() (K, V, error) {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

	if orderedMap.tail == nil {
		var nilKey K
		var nilValue V
		return nilKey, nilValue, errors.New("empty map")
	}
	node := orderedMap.tail
	delete(orderedMap.values, node.key)
	orderedMap.tail = node.prev
	if node.prev != nil {
		node.prev.next = nil
	} else {
		orderedMap.head = nil
	}
	return node.key, node.value, nil
}

func (orderedMap *OrderedMap[K, V]) RetrieveKey(key K) (V, error) {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

	node, ok := orderedMap.values[key]
	if !ok {
		var nilValue V
		return nilValue, errors.New("key not found")
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
	return node.value, nil
}

func (orderedMap *OrderedMap[K, V]) UpdateValue(key K, value V) error {
	orderedMap.mutex.Lock()
	defer orderedMap.mutex.Unlock()

	node, ok := orderedMap.values[key]
	if !ok {
		return errors.New("key not found")
	}
	node.value = value
	return nil
}

// returns the value associated with the provided key.
// returns an error if the key is not found.
func (orderedMap *OrderedMap[K, V]) GetValue(key K) (V, error) {
	orderedMap.mutex.RLock()
	defer orderedMap.mutex.RUnlock()

	if node, ok := orderedMap.values[key]; ok {
		return node.value, nil
	}
	var nilValue V
	return nilValue, errors.New("key not found")
}

// returns the keys in the order they were pushed
func (orderedMap *OrderedMap[K, V]) GetKeys() []K {
	orderedMap.mutex.RLock()
	defer orderedMap.mutex.RUnlock()

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
	orderedMap.mutex.RLock()
	defer orderedMap.mutex.RUnlock()

	values := make([]V, 0, len(orderedMap.values))
	currentNode := orderedMap.head
	for currentNode != nil {
		values = append(values, currentNode.value)
		currentNode = currentNode.next
	}
	return values
}
