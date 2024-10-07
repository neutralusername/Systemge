package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
)

type RequestResponseManager[T any] struct {
	config   *Config.RequestResponseManager
	requests map[string]*request[T]
	mutex    sync.RWMutex
}
type request[T any] struct {
	token           string
	responseChannel chan T
	doneChannel     chan struct{}
	responseLimit   uint64
	responseCount   uint64
}

func NewRequestResponseManager[T any](config *Config.RequestResponseManager) *RequestResponseManager[T] {
	if config == nil {
		config = &Config.RequestResponseManager{}
	}
	return &RequestResponseManager[T]{
		requests: make(map[string]*request[T]),
		mutex:    sync.RWMutex{},
		config:   config,
	}
}

// NewRequest creates a new request with the given token, response limit and timeout in milliseconds.
// If responseLimit is 0, it will be set to 1.
// If the token is too short or too long, an error will be returned.
// If the maximum number of active requests is reached, an error will be returned.
// If a request with the same token already exists, an error will be returned.
// If a timeout is set, the request will be aborted after the timeout.
// The request will be removed from the manager when the response limit is reached.
func (manager *RequestResponseManager[T]) NewRequest(token string, responseLimit uint64, timeoutMs uint64) (*request[T], error) {
	if responseLimit == 0 {
		responseLimit = 1
	}
	if manager.config.MinTokenLength > 0 && len(token) < manager.config.MinTokenLength {
		return nil, errors.New("token too short")
	}
	if manager.config.MaxTokenLength > 0 && len(token) > manager.config.MaxTokenLength {
		return nil, errors.New("token too long")
	}
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if manager.config.MaxActiveRequests > 0 && len(manager.requests) >= manager.config.MaxActiveRequests {
		return nil, errors.New("too many active requests")
	}
	if _, ok := manager.requests[token]; ok {
		return nil, errors.New("token already exists")
	}

	request := &request[T]{
		token:           token,
		responseChannel: make(chan T, responseLimit),
		doneChannel:     make(chan struct{}),
		responseLimit:   responseLimit,
		responseCount:   0,
	}
	manager.requests[token] = request

	if timeoutMs > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
				manager.AbortRequest(token)
			case <-request.doneChannel:
			}
		}()
	}

	return request, nil
}

// AddResponse adds a response to the request with the given token.
// If no request with the token exists, an error will be returned.
// If the response limit is reached, the request will be removed from the manager.
func (manager *RequestResponseManager[T]) AddResponse(token string, response T) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	request, ok := manager.requests[token]
	if !ok {
		return errors.New("no active request for token")
	}

	request.responseChannel <- response
	request.responseCount++

	if request.responseCount >= request.responseLimit {
		close(request.responseChannel)
		close(request.doneChannel)
		delete(manager.requests, token)
	}

	return nil
}

// AbortRequest aborts the request with the given token.
// If no request with the token exists, an error will be returned.
func (manager *RequestResponseManager[T]) AbortRequest(token string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	request, ok := manager.requests[token]
	if !ok {
		return errors.New("no active request for token")
	}

	close(request.doneChannel)
	close(request.responseChannel)
	delete(manager.requests, token)

	return nil
}

// GetRequest returns the request with the given token.
// If no request with the token exists, an error will be returned.
func (manager *RequestResponseManager[T]) GetRequest(token string) (*request[T], error) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	request, ok := manager.requests[token]
	if !ok {
		return nil, errors.New("no active request for token")
	}
	return request, nil
}

// GetActiveRequestTokens returns a list of all active request tokens.
func (manager *RequestResponseManager[T]) GetActiveRequestTokens() []string {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	tokens := make([]string, 0, len(manager.requests))
	for k := range manager.requests {
		tokens = append(tokens, k)
	}
	return tokens
}

// GetToken returns the token of the request.
func (request *request[T]) GetToken() string {
	return request.token
}

// GetResponseChannel returns the response channel of the request.
func (request *request[T]) GetResponseChannel() <-chan T {
	return request.responseChannel
}

// GetNextResponse returns the next response from the request.
// If the response channel is closed, an error will be returned.
// If the response channel is empty, it will block until a response is available.
func (request *request[T]) GetNextResponse() (T, error) {
	response, ok := <-request.responseChannel
	if !ok {
		var nilValue T
		return nilValue, errors.New("response channel closed")
	}
	return response, nil
}

// GetResponseCount returns the number of responses received.
func (request *request[T]) GetResponseCount() uint64 {
	return request.responseCount
}

// GetResponseLimit returns the response limit of the request.
func (request *request[T]) GetResponseLimit() uint64 {
	return request.responseLimit
}
