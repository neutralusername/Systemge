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
}

func NewTokenSemaphore(poolSize int, randomizerSeed int64) *TokenSemaphore {
	randomizer := NewRandomizer(randomizerSeed)
	tokens := make([]string, poolSize)
	for i := 0; i < poolSize; i++ {
		tokens[i] = randomizer.GenerateRandomString(18, ALPHA_NUMERIC)
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

func (s *TokenSemaphore) AcquireToken() string {
	token := <-s.channel
	s.mutex.Lock()
	s.acquiredTokens[token] = true
	s.mutex.Unlock()
	return token
}

func (s *TokenSemaphore) ReturnToken(token string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, exists := s.acquiredTokens[token]
	if !exists {
		return Error.New("Token is not valid", nil)
	}
	delete(s.acquiredTokens, token)
	s.channel <- s.randomizer.GenerateRandomString(18, ALPHA_NUMERIC)
	return nil
}
