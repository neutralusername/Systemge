package Tools

import (
	"errors"
	"sync"
)

type Pool[T comparable] struct {
	items map[T]bool // item -> isAvailable
	mutex sync.Mutex
}

// items must be comparable and unique.
// providing non comparable items, such as maps, slices, or functions, will result in a panic.
func NewPool[T comparable](maxItems uint32, initialItems []T) (*Pool[T], error) {
	if maxItems == 0 {
		return nil, errors.New("maxItems must be greater than 0")
	}
	if len(initialItems) > int(maxItems) {
		return nil, errors.New("initialItems must be less than or equal to maxItems")
	}
	pool := &Pool[T]{
		items:       make(map[T]bool),
		itemChannel: make(chan T, maxItems),
	}

	for _, item := range initialItems {
		if pool.items[item] {
			return nil, errors.New("duplicate item")
		}
		pool.items[item] = true
		pool.itemChannel <- item
	}

	return pool, nil
}

func (pool *Pool[T]) GetAcquiredItems() []T {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	acquiredItems := make([]T, 0)
	for item, isAvailable := range pool.items {
		if !isAvailable {
			acquiredItems = append(acquiredItems, item)
		}
	}

	return acquiredItems
}

func (pool *Pool[T]) GetAvailableItems() []T {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	availableItems := make([]T, 0)
	for item, isAvailable := range pool.items {
		if isAvailable {
			availableItems = append(availableItems, item)
		}
	}

	return availableItems
}

// GetItems returns a copy of the map of items.
func (pool *Pool[T]) GetItems() map[T]bool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	copiedItems := make(map[T]bool)
	for item, isAvailable := range pool.items {
		copiedItems[item] = isAvailable
	}
	return copiedItems
}

// AcquireItem returns an item from the pool.
// If the pool is empty, it will block until a item becomes available.
func (pool *Pool[T]) AcquireItem() T {
	item := <-pool.itemChannel
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.items[item] = false
	return item
}

// AcquireItemChannel returns a channel that will return an item from the pool.
// If the pool is empty, it will block until a item becomes available.
// The channel will be closed after the item is returned.
func (pool *Pool[T]) AcquireItemChannel() <-chan T {
	c := make(chan T, 1)
	go func() {
		item := <-pool.itemChannel
		pool.mutex.Lock()
		defer pool.mutex.Unlock()

		pool.items[item] = false
		c <- item
		close(c)
	}()
	return c
}

// ReturnItem returns an item from the pool.
// If the item does not exist, it will return an error.
// If the item is available, it will return a error.
func (pool *Pool[T]) ReturnItem(item T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	val, ok := pool.items[item]
	if !ok {
		return errors.New("item does not exist")
	}
	if val {
		return errors.New("item is available")
	}
	pool.items[item] = true
	pool.itemChannel <- item
	return nil
}

// ReplaceItem replaces an item in the pool.
// if returnItem is true, the item must be acquired.
// If the item does not exist, it will return an error.
// If the item is acquired, it will return an error.
// If the replacement already exists, it will return an error.
func (pool *Pool[T]) ReplaceItem(item T, replacement T, returnItem bool) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	val, ok := pool.items[item]
	if !ok {
		return errors.New("item does not exist")
	}
	if returnItem && !val {
		return errors.New("item is not acquired")
	}
	if replacement != item {
		if pool.items[replacement] {
			return errors.New("replacement already exists")
		}
		delete(pool.items, item)
	}
	pool.items[replacement] = true
	pool.itemChannel <- replacement
	return nil
}

// RemoveItems removes the item from the pool.
// if transactional is false, it will skip items that do not exist.
// if transactional is true, it will return an error if any item does not exist before modifying the pool.
func (pool *Pool[T]) RemoveItems(transactional bool, items ...T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if !transactional {
		for _, item := range items {
			if pool.items[item] {
				delete(pool.items, item)
			}
		}
		return nil
	} else {
		for _, item := range items {
			if !pool.items[item] {
				return errors.New("item does not exist")
			}
		}
		for _, item := range items {
			delete(pool.items, item)
		}
	}
	return nil
}

// AddItems adds new items to the pool.
// if transactional is false, it will skip items that already exist and stop when the pool is full.
// if transactional is true, it will return an error if any item already exists or if the amount of items exceeds the pool capacity.
func (pool *Pool[T]) AddItems(transactional bool, items ...T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if !transactional {
		for _, item := range items {
			if len(pool.items) == cap(pool.itemChannel) {
				break
			}
			if pool.items[item] {
				continue
			}
			pool.items[item] = true
			pool.itemChannel <- item
		}
	} else {
		if len(pool.items)+len(items) > cap(pool.itemChannel) {
			return errors.New("item count exceeds pool capacity")
		}
		for _, item := range items {
			if pool.items[item] {
				return errors.New("item already exists")
			}
		}
		for _, item := range items {
			pool.items[item] = true
			pool.itemChannel <- item
		}
	}
	return nil
}

// Clear removes all items from the pool and returns them.
func (pool *Pool[T]) Clear() map[T]bool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	items := make(map[T]bool)
	for item := range pool.items {
		items[item] = true
		delete(pool.items, item)
	}
	for len(pool.itemChannel) > 0 {
		<-pool.itemChannel
	}
	return items
}
