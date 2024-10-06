package Tools

import (
	"errors"
	"sync"
)

type AnySemaphore struct {
	items       map[any]bool // token -> isAvailable
	mutex       sync.Mutex
	itemChannel chan any
}

// items must be comparable and unique.
// providing non comparable items, such as maps, slices, or functions, will result in a panic.
func NewAnySemaphore(items []any) (*AnySemaphore, error) {
	tokenSemaphore := &AnySemaphore{
		items:       make(map[any]bool),
		itemChannel: make(chan any, len(items)),
	}

	for _, token := range items {
		if tokenSemaphore.items[token] {
			return nil, errors.New("duplicate token")
		}
		tokenSemaphore.items[token] = true
		tokenSemaphore.itemChannel <- token
	}

	return tokenSemaphore, nil
}

func (tokenSemaphore *AnySemaphore) GetAcquiredTokens() []any {
	tokenSemaphore.mutex.Lock()
	defer tokenSemaphore.mutex.Unlock()
	acquiredtokens := make([]any, 0)
	for token, isAvailable := range tokenSemaphore.items {
		if !isAvailable {
			acquiredtokens = append(acquiredtokens, token)
		}
	}

	return acquiredtokens
}

// AcquireToken returns a token from the pool.
// If the pool is empty, it will block until a token is available.
func (tokenSemaphore *AnySemaphore) AcquireToken() any {
	token := <-tokenSemaphore.itemChannel
	tokenSemaphore.mutex.Lock()
	defer tokenSemaphore.mutex.Unlock()
	tokenSemaphore.items[token] = false
	return token
}

// ReturnToken returns a token to the pool.
// If the token is not valid, it will return an error.
// replacementToken must be either same as token or a new token.
func (tokenSemaphore *AnySemaphore) ReturnToken(item any, replacementItem any) error {
	tokenSemaphore.mutex.Lock()
	defer tokenSemaphore.mutex.Unlock()
	if tokenSemaphore.items[item] {
		return errors.New("token is not acquired")
	}
	if replacementItem == "" {
		return errors.New("empty string token")
	}
	if replacementItem != item {
		if tokenSemaphore.items[replacementItem] {
			return errors.New("token already exists")
		}
		delete(tokenSemaphore.items, item)
	}
	tokenSemaphore.items[replacementItem] = true
	tokenSemaphore.itemChannel <- replacementItem
	return nil
}
