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
}

// items must be comparable and unique.
// providing non comparable items, such as maps, slices, or functions, will result in a panic.
func NewPool[T comparable](maxItems uint32, availableItems []T) (*Pool[T], error) {
	if maxItems == 0 {
		return nil, errors.New("maxItems must be greater than 0")
	}
	if len(availableItems) > int(maxItems) {
		return nil, errors.New("initialItems must be less than or equal to maxItems")
	}
	pool := &Pool[T]{
		acquiredItems:  make(map[T]bool),
		availableItems: make(map[T]bool),
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
	for item, isAvailable := range pool.acquiredItems {
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
	for item, isAvailable := range pool.availableItems {
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
	for item, isAvailable := range pool.acquiredItems {
		copiedItems[item] = isAvailable
	}
	return copiedItems
}

// AcquireItem returns an item from the pool.
// If the pool is empty, it will block until a item becomes available.
func (pool *Pool[T]) AcquireItem() T {
	pool.mutex.Lock()

	var item T
	if len(pool.availableItems) == 0 {
		waiter := make(chan T)
		pool.waiters = append(pool.waiters, waiter)
		pool.mutex.Unlock()
		item = <-waiter
	} else {
		for i := range pool.availableItems {
			pool.acquiredItems[item] = true
			delete(pool.availableItems, item)
			item = i
			break
		}
		pool.mutex.Unlock()
	}
	return item
}

func (pool *Pool[T]) TryAcquireItem() (T, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	var nilItem T
	if len(pool.availableItems) == 0 {
		return nilItem, errors.New("pool is empty")
	}
	for item := range pool.availableItems {
		pool.acquiredItems[item] = true
		delete(pool.availableItems, item)
		return item, nil
	}
	return nilItem, errors.New("can not occur")
}

// AcquireItemChannel returns a channel that will return an item from the pool.
// If the pool is empty, it will block until a item becomes available.
// The channel will be closed after the item is returned.
func (pool *Pool[T]) AcquireItemChannel() <-chan T {
	c := make(chan T, 1)
	go func() {
		pool.mutex.Lock()
		if len(pool.availableItems) == 0 {
			waiter := make(chan T)
			pool.waiters = append(pool.waiters, waiter)
			pool.mutex.Unlock()
			c <- <-waiter
		} else {
			for i := range pool.availableItems {
				pool.acquiredItems[i] = true
				delete(pool.availableItems, i)
				pool.mutex.Unlock()
				c <- i
				close(c)
				return
			}
		}
	}()
	return c
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
	if len(pool.waiters) > 0 {
		waiter := pool.waiters[0]
		pool.waiters = pool.waiters[1:]
		waiter <- item
	} else {
		pool.availableItems[item] = true
	}
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
		if len(pool.waiters) > 0 {
			waiter := pool.waiters[0]
			pool.waiters = pool.waiters[1:]
			waiter <- replacement
		} else {
			pool.availableItems[replacement] = true
		}
		return nil
	}
	if pool.availableItems[item] {
		delete(pool.availableItems, item)
		pool.availableItems[replacement] = true
		return nil
	}

	return errors.New("item does not exist")
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

// AddItems adds new items to the pool.
// if transactional is false, it will skip items that already exist and stop when the pool is full.
// if transactional is true, it will return an error if any item already exists or if the amount of items exceeds the pool capacity.
func (pool *Pool[T]) AddItems(transactional bool, items ...T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if !transactional {
		for _, item := range items {
			if pool.acquiredItems[item] || pool.availableItems[item] {
				continue
			}
			if len(pool.waiters) > 0 {
				waiter := pool.waiters[0]
				pool.waiters = pool.waiters[1:]
				waiter <- item
			} else {
				pool.availableItems[item] = true
			}
		}
	} else {
		for _, item := range items {
			if pool.acquiredItems[item] || pool.availableItems[item] {
				return errors.New("item already exists")
			}
		}
		for _, item := range items {
			if len(pool.availableItems)+len(pool.acquiredItems) >= len(pool.availableItems) {
				return errors.New("pool is full")
			}
			if len(pool.waiters) > 0 {
				waiter := pool.waiters[0]
				pool.waiters = pool.waiters[1:]
				waiter <- item
			} else {
				pool.availableItems[item] = true
			}
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
