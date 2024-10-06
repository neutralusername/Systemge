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
func NewGenericPool[T comparable](items []T) (*GenericPool[T], error) {
	genericPool := &GenericPool[T]{
		items:       make(map[T]bool),
		itemChannel: make(chan T, len(items)),
	}

	for _, item := range items {
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
func (genericPool *GenericPool[T]) ReturnItem(item T, replacementItem T) error {
	genericPool.mutex.Lock()
	defer genericPool.mutex.Unlock()
	if genericPool.items[item] {
		return errors.New("item is not acquired")
	}
	if replacementItem != item {
		if genericPool.items[replacementItem] {
			return errors.New("item already exists")
		}
		delete(genericPool.items, item)
	}
	genericPool.items[replacementItem] = true
	genericPool.itemChannel <- replacementItem
	return nil
}
