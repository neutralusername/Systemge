package Tools

import (
	"errors"
	"sync"
)

type GenericSemaphore[T comparable] struct {
	items       map[T]bool // item -> isAvailable
	mutex       sync.Mutex
	itemChannel chan T
}

// items must be comparable and unique.
// providing non comparable items, such as maps, slices, or functions, will result in a panic.
func NewGenericSemaphore[T comparable](items []T) (*GenericSemaphore[T], error) {
	genericSemaphore := &GenericSemaphore[T]{
		items:       make(map[T]bool),
		itemChannel: make(chan T, len(items)),
	}

	for _, item := range items {
		if genericSemaphore.items[item] {
			return nil, errors.New("duplicate item")
		}
		genericSemaphore.items[item] = true
		genericSemaphore.itemChannel <- item
	}

	return genericSemaphore, nil
}

func (genericSemaphore *GenericSemaphore[T]) GetAcquiredItems() []T {
	genericSemaphore.mutex.Lock()
	defer genericSemaphore.mutex.Unlock()
	acquiredItems := make([]T, 0)
	for item, isAvailable := range genericSemaphore.items {
		if !isAvailable {
			acquiredItems = append(acquiredItems, item)
		}
	}

	return acquiredItems
}

// AcquireItem returns a item from the pool.
// If the pool is empty, it will block until a item is available.
func (genericSemaphore *GenericSemaphore[T]) AcquireItem() T {
	item := <-genericSemaphore.itemChannel
	genericSemaphore.mutex.Lock()
	defer genericSemaphore.mutex.Unlock()
	genericSemaphore.items[item] = false
	return item
}

// ReturnItem returns a item to the pool.
// If the item is not valid, it will return an error.
// replacementItem must be either same as item or a new item.
func (genericSemaphore *GenericSemaphore[T]) ReturnItem(item T, replacementItem T) error {
	genericSemaphore.mutex.Lock()
	defer genericSemaphore.mutex.Unlock()
	if genericSemaphore.items[item] {
		return errors.New("item is not acquired")
	}
	if replacementItem != item {
		if genericSemaphore.items[replacementItem] {
			return errors.New("item already exists")
		}
		delete(genericSemaphore.items, item)
	}
	genericSemaphore.items[replacementItem] = true
	genericSemaphore.itemChannel <- replacementItem
	return nil
}
