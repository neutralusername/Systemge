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

// AcquireItem returns a item from the pool.
// If the pool is empty, it will block until a item is available.
func (genericPool *GenericPool[T]) AcquireItem() T {
	item := <-genericPool.itemChannel
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	genericPool.items[item] = false
	return item
}

// ReturnItem returns a item to the pool.
// If the item is not valid, it will return an error.
// replacementItem must be either same as item or a new item.
func (genericPool *GenericPool[T]) ReturnItem(item T, replacement T) error {
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

func (genericPool *GenericPool[T]) AddItem(item T) error {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	if genericPool.items[item] {
		return errors.New("item already exists")
	}
	if len(genericPool.itemChannel) == cap(genericPool.itemChannel) {
		return errors.New("pool is full")
	}
	genericPool.items[item] = true
	genericPool.itemChannel <- item
	return nil
}
