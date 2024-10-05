package Tools

import (
	"sync"
)

type TokenSemaphore struct {
	acquiredTokens  map[string]bool // token -> isAcquired
	availableTokens []string
	mutex           sync.Mutex
	randomizer      *Randomizer
	tokenSize       uint32
	tokenAlphabet   string
}

func NewTokenSemaphore(tokens []string) *TokenSemaphore {
	tokenSemaphore := &TokenSemaphore{
		acquiredTokens:  make(map[string]bool),
		tokenSize:       tokenSize,
		availableTokens: make([]string, 0, poolSize),
		randomizer:      NewRandomizer(randomizerSeed),
		tokenAlphabet:   tokenAlphabet,
	}

	for i := (uint32)(0); i < poolSize; i++ {
		token := tokenSemaphore.randomizer.GenerateRandomString(tokenSize, tokenAlphabet)
		for tokenSemaphore.acquiredTokens[token] {
			token = tokenSemaphore.randomizer.GenerateRandomString(tokenSize, tokenAlphabet)
		}
		tokenSemaphore.availableTokens = append(tokenSemaphore.availableTokens, token)
	}
	return tokenSemaphore
}

func (tokenSemaphore *TokenSemaphore) GetAcquiredTokens() []string {
	tokenSemaphore.mutex.Lock()
	defer tokenSemaphore.mutex.Unlock()
	acquiredTokens := make([]string, 0, len(tokenSemaphore.acquiredTokens))
	for token := range tokenSemaphore.acquiredTokens {
		acquiredTokens = append(acquiredTokens, token)
	}
	return acquiredTokens
}

// AcquireToken returns a token from the pool.
// If the pool is empty, it will block until a token is available.
func (tokenSemaphore *TokenSemaphore) AcquireToken() string {

}

// ReturnToken returns a token to the pool.
// If the token is not valid, it will return an error.
func (tokenSemaphore *TokenSemaphore) ReturnToken(token string) error {

}
