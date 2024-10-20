package tools

import (
	"errors"
	"sync"
)

type KeyDequeue[K comparable, V any] struct {
	head   *keyDequeueNode[K, V]
	tail   *keyDequeueNode[K, V]
	values map[K]*keyDequeueNode[K, V]
	mutex  sync.RWMutex
}

type keyDequeueNode[K comparable, V any] struct {
	key   K
	value V
	next  *keyDequeueNode[K, V]
	prev  *keyDequeueNode[K, V]
}

func NewKeyDequeue[K comparable, V any](initialItems map[K]V) *KeyDequeue[K, V] {
	orderedMap := &KeyDequeue[K, V]{
		values: make(map[K]*keyDequeueNode[K, V]),
	}
	for key, value := range initialItems {
		orderedMap.Push(key, value)
	}
	return orderedMap
}

func newKeyDequeueNode[K comparable, V any](key K, value V) *keyDequeueNode[K, V] {
	return &keyDequeueNode[K, V]{
		key:   key,
		value: value,
	}
}

// Push adds a new key-value pair to the end of the map.
// returns an error if the key already exists.
func (keyDequeue *KeyDequeue[K, V]) Push(key K, value V) error {
	keyDequeue.mutex.Lock()
	defer keyDequeue.mutex.Unlock()

	if _, ok := keyDequeue.values[key]; ok {
		return errors.New("key already exists")
	}

	node := newKeyDequeueNode(key, value)
	keyDequeue.values[key] = node
	if keyDequeue.head == nil {
		keyDequeue.head = node
		keyDequeue.tail = node
		return nil
	}

	keyDequeue.tail.next = node
	node.prev = keyDequeue.tail
	keyDequeue.tail = node
	return nil
}

// PopFront removes and returns the first key-value pair in the map.
// returns an error if the map is empty.
func (keyDequeue *KeyDequeue[K, V]) PopFront() (K, V, error) {
	keyDequeue.mutex.Lock()
	defer keyDequeue.mutex.Unlock()

	if keyDequeue.head == nil {
		var nilKey K
		var nilValue V
		return nilKey, nilValue, errors.New("empty map")
	}
	node := keyDequeue.head
	delete(keyDequeue.values, node.key)
	keyDequeue.head = node.next
	if node.next != nil {
		node.next.prev = nil
	} else {
		keyDequeue.tail = nil
	}
	return node.key, node.value, nil
}

// PopBack removes and returns the last key-value pair in the map.
// returns an error if the map is empty.
func (keyDequeue *KeyDequeue[K, V]) PopBack() (K, V, error) {
	keyDequeue.mutex.Lock()
	defer keyDequeue.mutex.Unlock()

	if keyDequeue.tail == nil {
		var nilKey K
		var nilValue V
		return nilKey, nilValue, errors.New("empty map")
	}
	node := keyDequeue.tail
	delete(keyDequeue.values, node.key)
	keyDequeue.tail = node.prev
	if node.prev != nil {
		node.prev.next = nil
	} else {
		keyDequeue.head = nil
	}
	return node.key, node.value, nil
}

// PopKey removes and returns the key-value pair associated with the provided key.
// returns an error if the key is not found.
func (keyDequeue *KeyDequeue[K, V]) PopKey(key K) (V, error) {
	keyDequeue.mutex.Lock()
	defer keyDequeue.mutex.Unlock()

	node, ok := keyDequeue.values[key]
	if !ok {
		var nilValue V
		return nilValue, errors.New("key not found")
	}
	delete(keyDequeue.values, key)
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		keyDequeue.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		keyDequeue.tail = node.prev
	}
	return node.value, nil
}

// Update updates the value associated with the provided key.
// returns an error if the key is not found.
func (keyDequeue *KeyDequeue[K, V]) Update(key K, value V) error {
	keyDequeue.mutex.Lock()
	defer keyDequeue.mutex.Unlock()

	node, ok := keyDequeue.values[key]
	if !ok {
		return errors.New("key not found")
	}
	node.value = value
	return nil
}

// Upsert inserts or updates the value associated with the provided key.
func (keyDequeue *KeyDequeue[K, V]) Upsert(key K, value V) {
	keyDequeue.mutex.Lock()
	defer keyDequeue.mutex.Unlock()

	if node, ok := keyDequeue.values[key]; ok {
		node.value = value
		return
	}
	keyDequeue.Push(key, value)
}

// returns the value associated with the provided key.
// returns an error if the key is not found.
func (keyDequeue *KeyDequeue[K, V]) GetValue(key K) (V, error) {
	keyDequeue.mutex.RLock()
	defer keyDequeue.mutex.RUnlock()

	if node, ok := keyDequeue.values[key]; ok {
		return node.value, nil
	}
	var nilValue V
	return nilValue, errors.New("key not found")
}

// returns the keys in the order they were pushed
func (keyDequeue *KeyDequeue[K, V]) GetKeys() []K {
	keyDequeue.mutex.RLock()
	defer keyDequeue.mutex.RUnlock()

	keys := make([]K, 0, len(keyDequeue.values))
	currentNode := keyDequeue.head
	for currentNode != nil {
		keys = append(keys, currentNode.key)
		currentNode = currentNode.next
	}
	return keys
}

// returns the values in the order they were pushed
func (keyDequeue *KeyDequeue[K, V]) GetValues() []V {
	keyDequeue.mutex.RLock()
	defer keyDequeue.mutex.RUnlock()

	values := make([]V, 0, len(keyDequeue.values))
	currentNode := keyDequeue.head
	for currentNode != nil {
		values = append(values, currentNode.value)
		currentNode = currentNode.next
	}
	return values
}

// returns the number of elements in the key-dequeue
func (keyDequeue *KeyDequeue[K, V]) Len() int {
	keyDequeue.mutex.RLock()
	defer keyDequeue.mutex.RUnlock()

	return len(keyDequeue.values)
}
