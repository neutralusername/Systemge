package Tools

import (
	"errors"
	"sync"
)

type GenericPool[T comparable] struct {
	items       map[T]bool // item -> isAvailable
	mutex       sync.Mutex
	itemChannel chan T
}

// items must be comparable and unique.
// providing non comparable items, such as maps, slices, or functions, will result in a panic.
func NewGenericPool[T comparable](maxItems uint32, initialItems []T) (*GenericPool[T], error) {
	if maxItems == 0 {
		return nil, errors.New("maxItems must be greater than 0")
	}
	if len(initialItems) > int(maxItems) {
		return nil, errors.New("initialItems must be less than or equal to maxItems")
	}
	genericPool := &GenericPool[T]{
		items:       make(map[T]bool),
		itemChannel: make(chan T, maxItems),
	}

	for _, item := range initialItems {
		if genericPool.items[item] {
			return nil, errors.New("duplicate item")
		}
		genericPool.items[item] = true
		genericPool.itemChannel <- item
	}

	return genericPool, nil
}

func (genericPool *GenericPool[T]) GetAcquiredItems() []T {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	acquiredItems := make([]T, 0)
	for item, isAvailable := range genericPool.items {
		if !isAvailable {
			acquiredItems = append(acquiredItems, item)
		}
	}

	return acquiredItems
}

func (genericPool *GenericPool[T]) GetAvailableItems() []T {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	availableItems := make([]T, 0)
	for item, isAvailable := range genericPool.items {
		if isAvailable {
			availableItems = append(availableItems, item)
		}
	}

	return availableItems
}

// GetItems returns a copy of the map of items.
func (genericPool *GenericPool[T]) GetItems() map[T]bool {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	copiedItems := make(map[T]bool)
	for item, isAvailable := range genericPool.items {
		copiedItems[item] = isAvailable
	}
	return copiedItems
}

// AcquireItem returns an item from the pool.
// If the pool is empty, it will block until a item is available.
func (genericPool *GenericPool[T]) AcquireItem() T {
	item := <-genericPool.itemChannel
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	genericPool.items[item] = false
	return item
}

// TryAcquireItem returns an item from the pool.
// If the item does not exist, it will return an error.
// If the item is available, it will return a error.
func (genericPool *GenericPool[T]) ReturnItem(item T) error {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	val, ok := genericPool.items[item]
	if !ok {
		return errors.New("item does not exist")
	}
	if val {
		return errors.New("item is available")
	}
	genericPool.items[item] = true
	genericPool.itemChannel <- item
	return nil
}

// ReplaceItem replaces an item in the pool.
// If the item does not exist, it will return an error.
// If the item is acquired, it will return an error.
// If the replacement already exists, it will return an error.
func (genericPool *GenericPool[T]) ReplaceItem(item T, replacement T) error {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	val, ok := genericPool.items[item]
	if !ok {
		return errors.New("item does not exist")
	}
	if !val {
		return errors.New("item is not acquired")
	}
	if replacement != item {
		if genericPool.items[replacement] {
			return errors.New("replacement already exists")
		}
		delete(genericPool.items, item)
	}
	genericPool.items[replacement] = true
	genericPool.itemChannel <- replacement
	return nil
}

// RemoveItem removes a item from the pool.
// If the item does not exist, it will return an error.
func (genericPool *GenericPool[T]) RemoveItem(item T) error {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	_, ok := genericPool.items[item]
	if !ok {
		return errors.New("item does not exist")
	}
	delete(genericPool.items, item)
	return nil
}

// AddItem adds a new item to the pool.
// If the item already exists, it will return an error.
// If the pool is full, it will return an error.
func (genericPool *GenericPool[T]) AddItems(item ...T) error {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	if len(genericPool.items)+len(item) > cap(genericPool.itemChannel) {
		return errors.New("item count exceeds pool capacity")
	}
	for _, item := range item {
		if genericPool.items[item] {
			return errors.New("item already exists")
		}
		genericPool.items[item] = true
		genericPool.itemChannel <- item
	}
	return nil
}

// Clear removes all items from the pool and returns them.
func (genericPool *GenericPool[T]) Clear() []T {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	for item := range genericPool.items {
		delete(genericPool.items, item)
	}
	items := make([]T, 0)
	for len(genericPool.itemChannel) > 0 {
		item := <-genericPool.itemChannel
		items = append(items, item)
	}
	return items
}
