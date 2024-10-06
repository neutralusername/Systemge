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
	itemPool := &GenericPool[T]{
		items:       make(map[T]bool),
		itemChannel: make(chan T, len(items)),
	}

	for _, item := range items {
		if itemPool.items[item] {
			return nil, errors.New("duplicate item")
		}
		itemPool.items[item] = true
		itemPool.itemChannel <- item
	}

	return itemPool, nil
}

func (itemPool *GenericPool[T]) GetAcquiredItems() []T {
	itemPool.mutex.Lock()
	defer itemPool.mutex.Unlock()
	acquiredItems := make([]T, 0)
	for item, isAvailable := range itemPool.items {
		if !isAvailable {
			acquiredItems = append(acquiredItems, item)
		}
	}

	return acquiredItems
}

func (itemPool *GenericPool[T]) GetAvailableItems() []T {
	itemPool.mutex.Lock()
	defer itemPool.mutex.Unlock()
	availableItems := make([]T, 0)
	for item, isAvailable := range itemPool.items {
		if isAvailable {
			availableItems = append(availableItems, item)
		}
	}

	return availableItems
}

// GetItems returns a copy of the map of items.
func (itemPool *GenericPool[T]) GetItems() map[T]bool {
	itemPool.mutex.Lock()
	defer itemPool.mutex.Unlock()
	copiedItems := make(map[T]bool)
	for item, isAvailable := range itemPool.items {
		copiedItems[item] = isAvailable
	}
	return copiedItems
}

// AcquireItem returns a item from the pool.
// If the pool is empty, it will block until a item is available.
func (itemPool *GenericPool[T]) AcquireItem() T {
	item := <-itemPool.itemChannel
	itemPool.mutex.Lock()
	defer itemPool.mutex.Unlock()
	itemPool.items[item] = false
	return item
}

// ReturnItem returns a item to the pool.
// If the item is not valid, it will return an error.
// replacementItem must be either same as item or a new item.
func (itemPool *GenericPool[T]) ReturnItem(item T, replacementItem T) error {
	itemPool.mutex.Lock()
	defer itemPool.mutex.Unlock()
	if itemPool.items[item] {
		return errors.New("item is not acquired")
	}
	if replacementItem != item {
		if itemPool.items[replacementItem] {
			return errors.New("item already exists")
		}
		delete(itemPool.items, item)
	}
	itemPool.items[replacementItem] = true
	itemPool.itemChannel <- replacementItem
	return nil
}
