package Tools

import (
	"errors"
	"sync"
)

type Pool[T comparable] struct {
	acquiredItems  map[T]bool // item -> isAvailable
	availableItems map[T]bool
	mutex          sync.Mutex
	waiters        []chan T
	maxItems       uint32
}

// items must be comparable and unique.
// providing non comparable items, such as maps, slices, or functions, will result in a panic.
// maxItems == 0 means no limit.
func NewPool[T comparable](maxItems uint32, availableItems []T) (*Pool[T], error) {
	if maxItems > 0 && len(availableItems) > int(maxItems) {
		return nil, errors.New("initialItems must be less than or equal to maxItems")
	}
	pool := &Pool[T]{
		acquiredItems:  make(map[T]bool),
		availableItems: make(map[T]bool),
		maxItems:       maxItems,
	}

	for _, item := range availableItems {
		if pool.availableItems[item] {
			return nil, errors.New("duplicate item")
		}
		pool.availableItems[item] = true
	}

	return pool, nil
}

func (pool *Pool[T]) GetAcquiredItems() []T {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	acquiredItems := make([]T, 0)
	for item := range pool.acquiredItems {
		acquiredItems = append(acquiredItems, item)
	}

	return acquiredItems
}

func (pool *Pool[T]) GetAvailableItems() []T {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	availableItems := make([]T, 0)
	for item := range pool.availableItems {
		availableItems = append(availableItems, item)
	}

	return availableItems
}

// GetItems returns a map of items in the pool. The value is true if the item is available.
func (pool *Pool[T]) GetItems() map[T]bool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	items := make(map[T]bool)
	for item, _ := range pool.acquiredItems {
		items[item] = false
	}
	for item, _ := range pool.availableItems {
		items[item] = true
	}
	return items
}

// AcquireItem returns an item from the pool.
// If the pool is empty, it will block until a item becomes available.
func (pool *Pool[T]) AcquireItem() T {
	pool.mutex.Lock()

	for item := range pool.availableItems {
		pool.acquiredItems[item] = true
		delete(pool.availableItems, item)
		pool.mutex.Unlock()
		return item
	}

	waiter := make(chan T)
	pool.waiters = append(pool.waiters, waiter)
	pool.mutex.Unlock()
	return <-waiter
}

func (pool *Pool[T]) TryAcquireItem() (T, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	for item := range pool.availableItems {
		pool.acquiredItems[item] = true
		delete(pool.availableItems, item)
		return item, nil
	}

	var nilItem T
	return nilItem, errors.New("pool is empty")
}

// AcquireItemChannel returns a channel that will return an item from the pool.
// If the pool is empty, it will block until a item becomes available.
// The channel will be closed after the item is returned.
func (pool *Pool[T]) AcquireItemChannel() <-chan T {
	c := make(chan T, 1)
	go func() {
		c <- pool.AcquireItem()
		close(c)
	}()
	return c
}

// RemoveItems removes the item from the pool.
// if transactional is false, it will skip items that do not exist.
// if transactional is true, it will return an error if any item does not exist before modifying the pool.
func (pool *Pool[T]) RemoveItems(transactional bool, items ...T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if !transactional {
		for _, item := range items {
			if pool.acquiredItems[item] {
				delete(pool.acquiredItems, item)
			}
			if pool.availableItems[item] {
				delete(pool.availableItems, item)
			}
		}
	} else {
		for _, item := range items {
			if !pool.acquiredItems[item] && !pool.availableItems[item] {
				return errors.New("item does not exist")
			}
		}
		for _, item := range items {
			delete(pool.acquiredItems, item)
			delete(pool.availableItems, item)
		}
	}
	return nil
}

// Clear removes all items from the pool and returns them.
func (pool *Pool[T]) Clear() map[T]bool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	items := make(map[T]bool)
	for item, isAvailable := range pool.acquiredItems {
		if isAvailable {
			pool.availableItems[item] = true
		} else {
			items[item] = true
		}
	}
	pool.acquiredItems = make(map[T]bool)
	return items
}

// ReturnItem returns an item from the pool.
// If the item does not exist, it will return an error.
// If the item is available, it will return a error.
func (pool *Pool[T]) ReturnItem(item T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if !pool.acquiredItems[item] {
		return errors.New("item does not exist")
	}

	delete(pool.acquiredItems, item)
	pool.addItem(item)
	return nil
}

// ReplaceItem replaces an item in the pool.
// If the item does not exist, it will return an error.
// If the item is acquired, it will return an error.
// If the replacement already exists, it will return an error.
func (pool *Pool[T]) ReplaceItem(item T, replacement T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if item != replacement {
		if pool.acquiredItems[replacement] {
			return errors.New("replacement already exists")
		}
		if pool.availableItems[replacement] {
			return errors.New("replacement already exists")
		}
	}

	if pool.acquiredItems[item] {
		delete(pool.acquiredItems, item)
		pool.addItem(replacement)
		return nil
	}
	if pool.availableItems[item] {
		delete(pool.availableItems, item)
		pool.addItem(replacement)
		return nil
	}

	return errors.New("item does not exist")
}

// AddItems adds new items to the pool.
// if transactional is false, it will skip items that already exist and stop when the pool is full.
// if transactional is true, it will return an error if any item already exists or if the amount of items exceeds the pool capacity.
func (pool *Pool[T]) AddItems(transactional bool, items ...T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if !transactional {
		for _, item := range items {
			if pool.maxItems > 0 && len(pool.availableItems)+len(pool.acquiredItems) == len(pool.availableItems) {
				break
			}
			if pool.acquiredItems[item] || pool.availableItems[item] {
				continue
			}
			pool.addItem(item)
		}
	} else {
		if pool.maxItems > 0 && len(pool.availableItems)+len(pool.acquiredItems)+len(items) > len(pool.availableItems) {
			return errors.New("amount of items exceeds pool capacity")
		}
		for _, item := range items {
			if pool.acquiredItems[item] || pool.availableItems[item] {
				return errors.New("an item already exists")
			}
		}
		for _, item := range items {
			pool.addItem(item)
		}
	}
	return nil
}

func (pool *Pool[T]) addItem(item T) {
	if len(pool.waiters) > 0 {
		waiter := pool.waiters[0]
		pool.waiters = pool.waiters[1:]
		waiter <- item
	} else {
		pool.availableItems[item] = true
	}
}
