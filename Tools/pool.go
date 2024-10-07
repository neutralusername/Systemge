package Tools

import (
	"errors"
	"sync"
	"time"
)

type Pool[T comparable] struct {
	items    map[T]bool // item -> isAvailable
	mutex    sync.Mutex
	waiters  map[chan T]bool
	maxItems uint32
}

// items must be comparable and unique.
// providing non comparable items, such as maps, slices, or functions, will result in a panic.
// maxItems == 0 means no limit.
func NewPool[T comparable](maxItems uint32, availableItems []T) (*Pool[T], error) {
	if maxItems > 0 && len(availableItems) > int(maxItems) {
		return nil, errors.New("initialItems must be less than or equal to maxItems")
	}
	pool := &Pool[T]{
		items:    make(map[T]bool),
		waiters:  make(map[chan T]bool),
		maxItems: maxItems,
	}

	for _, item := range availableItems {
		if pool.items[item] {
			return nil, errors.New("duplicate item")
		}
		pool.items[item] = true
	}

	return pool, nil
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

// GetItems returns a map of items in the pool. The value is true if the item is available.
func (pool *Pool[T]) GetItems() map[T]bool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	items := make(map[T]bool)
	for item, isAvailable := range pool.items {
		items[item] = isAvailable
	}
	return items
}

// AcquireItem returns an item from the pool.
// If the pool is empty, it will block until a item becomes available.
func (pool *Pool[T]) AcquireItem(timeout uint32) (T, error) {
	pool.mutex.Lock()

	for item, isAvailable := range pool.items {
		if isAvailable {
			pool.items[item] = false
			pool.mutex.Unlock()
			return item, nil
		}
	}

	waiter := make(chan T)
	pool.waiters[waiter] = true
	pool.mutex.Unlock()

	if timeout == 0 {
		return <-waiter, nil
	} else {
		select {
		case item := <-waiter:
			return item, nil
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			pool.mutex.Lock()
			defer pool.mutex.Unlock()
			delete(pool.waiters, waiter)
			var nilItem T
			return nilItem, errors.New("timeout")
		}
	}
}

func (pool *Pool[T]) TryAcquireItem() (T, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	for item, isAvailable := range pool.items {
		if isAvailable {
			pool.items[item] = false
			return item, nil
		}
	}

	var nilItem T
	return nilItem, errors.New("no items available")
}

// AcquireItemChannel returns a channel that will return an item from the pool.
// If the pool is empty, it will block until a item becomes available.
// The channel will be closed after the item is returned.
func (pool *Pool[T]) AcquireItemChannel(timeoutMs uint32) <-chan T {
	channel := make(chan T, 1)
	go func() {
		item, err := pool.AcquireItem(timeoutMs)
		if err != nil {
			close(channel)
		} else {
			channel <- item
			close(channel)
		}
	}()
	return channel
}

// RemoveItems removes the item from the pool.
// if transactional is true, it will either remove all items or none.
func (pool *Pool[T]) RemoveItems(transactional bool, items ...T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if transactional {
		for _, item := range items {
			if !pool.items[item] {
				return errors.New("item does not exist")
			}
		}
	}
	for _, item := range items {
		delete(pool.items, item)
	}
	return nil
}

// Clear removes all items from the pool and returns them.
func (pool *Pool[T]) Clear() map[T]bool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	items := pool.items
	pool.items = make(map[T]bool)
	return items
}

// ReturnItem returns an item from the pool.
// If the item does not exist, it will return an error.
// If the item is available, it will return a error.
func (pool *Pool[T]) ReturnItem(item T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	isAvailable, ok := pool.items[item]
	if !ok {
		return errors.New("item does not exist")
	}
	if isAvailable {
		return errors.New("item is not acquired")
	}
	delete(pool.items, item)
	pool.addItem(item)
	return nil
}

// ReplaceItem replaces an item in the pool.
// If the item does not exist, it will return an error.
// If isReturned is true, the item must be acquired.
// If the replacement already exists, it will return an error.
func (pool *Pool[T]) ReplaceItem(item T, replacement T, isReturned bool) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	isAvailable, ok := pool.items[item]
	if !ok {
		return errors.New("item does not exist")
	}
	if isAvailable && isReturned {
		return errors.New("item is not acquired")
	}
	if pool.items[replacement] {
		return errors.New("replacement already exists")
	}

	delete(pool.items, item)
	pool.addItem(replacement)
	return nil
}

// AddItems adds new items to the pool.
// if transactional is true, it will either add all items or none.
func (pool *Pool[T]) AddItems(transactional bool, items ...T) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if transactional {
		if pool.maxItems > 0 && len(pool.items)+len(items) > int(pool.maxItems) {
			return errors.New("amount of items exceeds pool capacity")
		}
		for _, item := range items {
			if pool.items[item] {
				return errors.New("an item already exists")
			}
		}
	}
	for _, item := range items {
		if pool.maxItems > 0 && len(pool.items) == int(pool.maxItems) {
			break
		}
		if pool.items[item] {
			continue
		}
		pool.addItem(item)
	}
	return nil
}

func (pool *Pool[T]) addItem(item T) {
	if len(pool.waiters) > 0 {
		var waiter chan T
		for waiter = range pool.waiters {
			delete(pool.waiters, waiter)
			break
		}
		waiter <- item
	} else {
		pool.items[item] = true
	}
}
