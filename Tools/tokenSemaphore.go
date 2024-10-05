package Tools

type TokenSemaphore struct {
	tokens       map[string]bool // token -> isAvailable
	tokenChannel chan string
}

// always uses the provided tokens. If the pool is empty, it will block until a token is available.
// tokens can be returned manually.
func NewTokenSemaphore(tokens []string) *TokenSemaphore {
	tokenSemaphore := &TokenSemaphore{
		tokens:       make(map[string]bool),
		tokenChannel: make(chan string, len(tokens)),
	}

	for _, token := range tokens {
		tokenSemaphore.tokens[token] = true
		tokenSemaphore.tokenChannel <- token
	}

	return tokenSemaphore
}

func (tokenSemaphore *TokenSemaphore) GetAcquiredTokens() []string {
	acquiredtokens := make([]string, 0)
	for token, isAvailable := range tokenSemaphore.tokens {
		if !isAvailable {
			acquiredtokens = append(acquiredtokens, token)
		}
	}
	return acquiredtokens
}

// AcquireToken returns a token from the pool.
// If the pool is empty, it will block until a token is available.
func (tokenSemaphore *TokenSemaphore) AcquireToken() string {
	token := <-tokenSemaphore.tokenChannel
	tokenSemaphore.tokens[token] = false
	return token
}

// ReturnToken returns a token to the pool.
// If the token is not valid, it will return an error.
func (tokenSemaphore *TokenSemaphore) ReturnToken(token string, replacementToken string) error {
	if tokenSemaphore.tokens[token] {
		return nil
	}
	delete(tokenSemaphore.tokens, token)
	tokenSemaphore.tokens[replacementToken] = true
	tokenSemaphore.tokenChannel <- replacementToken
	return nil
}
