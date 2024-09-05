package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Error"
)

type TokenSemaphore struct {
	tokens  map[string]bool // token -> isAcquired
	channel chan string
	mutex   sync.Mutex
}

func NewTokenSemaphore(pool []string) *TokenSemaphore {
	tokenMap := make(map[string]bool)
	for _, token := range pool {
		tokenMap[token] = false
	}
	channel := make(chan string, len(pool))
	for _, token := range pool {
		channel <- token
	}
	return &TokenSemaphore{
		tokens:  tokenMap,
		channel: channel,
	}
}

func (s *TokenSemaphore) AcquireToken() string {
	token := <-s.channel
	s.mutex.Lock()
	s.tokens[token] = true
	s.mutex.Unlock()
	return token
}

func (s *TokenSemaphore) ReturnToken(token string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	isAcquired, exists := s.tokens[token]
	if !exists {
		return Error.New("Token is not valid", nil)
	}
	if !isAcquired {
		return Error.New("Token is already returned", nil)
	}
	s.tokens[token] = false
	s.channel <- token
	return nil
}
