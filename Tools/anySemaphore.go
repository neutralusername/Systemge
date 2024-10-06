package Tools

import (
	"errors"
	"sync"
)

type AnySemaphore struct {
	items       map[any]bool // item -> isAvailable
	mutex       sync.Mutex
	itemChannel chan any
}

// items must be comparable and unique.
// providing non comparable items, such as maps, slices, or functions, will result in a panic.
func NewAnySemaphore(items []any) (*AnySemaphore, error) {
	anySemaphore := &AnySemaphore{
		items:       make(map[any]bool),
		itemChannel: make(chan any, len(items)),
	}

	for _, item := range items {
		if anySemaphore.items[item] {
			return nil, errors.New("duplicate item")
		}
		anySemaphore.items[item] = true
		anySemaphore.itemChannel <- item
	}

	return anySemaphore, nil
}

func (anySemaphore *AnySemaphore) GetAcquiredItems() []any {
	anySemaphore.mutex.Lock()
	defer anySemaphore.mutex.Unlock()
	acquiredItems := make([]any, 0)
	for item, isAvailable := range anySemaphore.items {
		if !isAvailable {
			acquiredItems = append(acquiredItems, item)
		}
	}

	return acquiredItems
}

// AcquireItem returns a item from the pool.
// If the pool is empty, it will block until a item is available.
func (anySemaphore *AnySemaphore) AcquireItem() any {
	item := <-anySemaphore.itemChannel
	anySemaphore.mutex.Lock()
	defer anySemaphore.mutex.Unlock()
	anySemaphore.items[item] = false
	return item
}

// ReturnItem returns a item to the pool.
// If the item is not valid, it will return an error.
// replacementItem must be either same as item or a new item.
func (anySemaphore *AnySemaphore) ReturnItem(item any, replacementItem any) error {
	anySemaphore.mutex.Lock()
	defer anySemaphore.mutex.Unlock()
	if anySemaphore.items[item] {
		return errors.New("item is not acquired")
	}
	if replacementItem == "" {
		return errors.New("empty string item")
	}
	if replacementItem != item {
		if anySemaphore.items[replacementItem] {
			return errors.New("item already exists")
		}
		delete(anySemaphore.items, item)
	}
	anySemaphore.items[replacementItem] = true
	anySemaphore.itemChannel <- replacementItem
	return nil
}
