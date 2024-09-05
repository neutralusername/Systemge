package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
)

type TokenSemaphore struct {
	acquiredTokens map[string]bool // token -> isAcquired
	channel        chan string
	mutex          sync.Mutex
}

func NewTokenSemaphore(poolSize int) *TokenSemaphore {
	tokens := make([]string, poolSize)
	for i := 0; i < poolSize; i++ {
		tokens[i] = GenerateRandomString(18, Tools.ALPHA_NUMERIC)
	}
	channel := make(chan string, poolSize)
	for _, token := range tokens {
		channel <- token
	}
	return &TokenSemaphore{
		acquiredTokens: make(map[string]bool),
		channel:        channel,
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
	s.channel <- GenerateRandomString(18, Tools.ALPHA_NUMERIC)
	return nil
}
