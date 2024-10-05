package Tools

import (
	"errors"
	"sync"
)

type TokenSemaphore struct {
	tokens       map[string]bool // token -> isAvailable
	mutex        sync.Mutex
	tokenChannel chan string
}

// always uses the provided tokens. If the pool is empty, it will block until a token is available.
// tokens can be returned manually.
func NewTokenSemaphore(tokens []string) (*TokenSemaphore, error) {
	tokenSemaphore := &TokenSemaphore{
		tokens:       make(map[string]bool),
		tokenChannel: make(chan string, len(tokens)),
	}

	for _, token := range tokens {
		if token == "" {
			return nil, errors.New("empty string token")
		}
		tokenSemaphore.tokens[token] = true
		tokenSemaphore.tokenChannel <- token
	}

	return tokenSemaphore, nil
}

func (tokenSemaphore *TokenSemaphore) GetAcquiredTokens() []string {
	tokenSemaphore.mutex.Lock()
	acquiredtokens := make([]string, 0)
	for token, isAvailable := range tokenSemaphore.tokens {
		if !isAvailable {
			acquiredtokens = append(acquiredtokens, token)
		}
	}
	tokenSemaphore.mutex.Unlock()
	return acquiredtokens
}

// AcquireToken returns a token from the pool.
// If the pool is empty, it will block until a token is available.
func (tokenSemaphore *TokenSemaphore) AcquireToken() string {
	token := <-tokenSemaphore.tokenChannel
	tokenSemaphore.mutex.Lock()
	tokenSemaphore.tokens[token] = false
	tokenSemaphore.mutex.Unlock()
	return token
}

// ReturnToken returns a token to the pool.
// If the token is not valid, it will return an error.
// replacementToken must be either same as token or a new token.
func (tokenSemaphore *TokenSemaphore) ReturnToken(token string, replacementToken string) error {
	tokenSemaphore.mutex.Lock()
	defer tokenSemaphore.mutex.Unlock()
	if tokenSemaphore.tokens[token] {
		return errors.New("token is not acquired")
	}
	if replacementToken == "" {
		return errors.New("empty string token")
	}
	if replacementToken != token {
		if tokenSemaphore.tokens[replacementToken] {
			return errors.New("token already exists")
		}
		delete(tokenSemaphore.tokens, token)
	}
	tokenSemaphore.tokens[replacementToken] = true
	tokenSemaphore.tokenChannel <- replacementToken
	return nil
}
