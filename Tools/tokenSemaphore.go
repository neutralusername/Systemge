package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Error"
)

type TokenSemaphore struct {
	acquiredTokens map[string]bool // token -> isAcquired
	channel        chan string
	mutex          sync.Mutex
	randomizer     *Randomizer
	tokenSize      uint32
}

func NewTokenSemaphore(poolSize int, tokenSize uint32, randomizerSeed int64) *TokenSemaphore {
	if poolSize <= 0 {
		panic("Pool size must be greater than 0")
	}
	randomizer := NewRandomizer(randomizerSeed)
	tokens := make([]string, poolSize)
	for i := 0; i < poolSize; i++ {
		tokens[i] = randomizer.GenerateRandomString(tokenSize, ALPHA_NUMERIC)
	}
	channel := make(chan string, poolSize)
	for _, token := range tokens {
		channel <- token
	}
	return &TokenSemaphore{
		acquiredTokens: make(map[string]bool),
		channel:        channel,
		randomizer:     randomizer,
	}
}

func (tokenSemaphore *TokenSemaphore) AcquireToken() string {
	token := <-tokenSemaphore.channel
	tokenSemaphore.mutex.Lock()
	tokenSemaphore.acquiredTokens[token] = true
	tokenSemaphore.mutex.Unlock()
	return token
}

func (tokenSemaphore *TokenSemaphore) ReturnToken(token string) error {
	tokenSemaphore.mutex.Lock()
	defer tokenSemaphore.mutex.Unlock()
	_, exists := tokenSemaphore.acquiredTokens[token]
	if !exists {
		return Error.New("Token is not valid", nil)
	}
	delete(tokenSemaphore.acquiredTokens, token)
	tokenSemaphore.channel <- tokenSemaphore.randomizer.GenerateRandomString(tokenSemaphore.tokenSize, ALPHA_NUMERIC)
	return nil
}
