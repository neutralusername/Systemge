package tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/systemge/configs"
)

type OnResponse[T any] func(*Request[T], T)

type RequestResponseManager[T any] struct {
	config   *configs.RequestResponseManager
	requests map[string]*Request[T]
	mutex    sync.RWMutex
}
type Request[T any] struct {
	token           string
	responseChannel chan T
	doneChannel     chan struct{}
	responseLimit   uint64
	responseCount   uint64
	onResponse      OnResponse[T]
}

func NewRequestResponseManager[T any](config *configs.RequestResponseManager) *RequestResponseManager[T] {
	if config == nil {
		config = &configs.RequestResponseManager{}
	}
	return &RequestResponseManager[T]{
		requests: make(map[string]*Request[T]),
		mutex:    sync.RWMutex{},
		config:   config,
	}
}

// NewRequest creates a new request with the given token, response limit and timeout in nanoseconds.
// If responseLimit is 0, it will be set to 1.
// If the token is too short or too long, an error will be returned.
// If the maximum number of active requests is reached, an error will be returned.
// If a request with the same token already exists, an error will be returned.
// If a timeout is set, the request will be aborted after the timeout.
// The request will be removed from the manager when the response limit is reached.
func (manager *RequestResponseManager[T]) NewRequest(token string, responseLimit uint64, timeoutNs int64, onResponse OnResponse[T]) (*Request[T], error) {
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

	request := &Request[T]{
		token:           token,
		responseChannel: make(chan T, responseLimit),
		doneChannel:     make(chan struct{}),
		responseLimit:   responseLimit,
		responseCount:   0,
		onResponse:      onResponse,
	}
	manager.requests[token] = request

	if timeoutNs > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(timeoutNs) * time.Nanosecond):
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

	if request.onResponse != nil {
		go request.onResponse(request, response)
	}

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
func (manager *RequestResponseManager[T]) GetRequest(token string) (*Request[T], error) {
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
func (request *Request[T]) GetToken() string {
	return request.token
}

// GetResponseChannel returns the response channel of the request.
func (request *Request[T]) GetResponseChannel() <-chan T {
	return request.responseChannel
}

// GetNextResponse returns the next response from the request.
// If the response channel is closed, an error will be returned.
// If the response channel is empty, it will block until a response is available.
func (request *Request[T]) GetNextResponse() (T, error) {
	response, ok := <-request.responseChannel
	if !ok {
		var nilValue T
		return nilValue, errors.New("response channel closed")
	}
	return response, nil
}

// GetResponses returns all responses received.
func (request *Request[T]) GetResponses() []T {
	responses := make([]T, 0, request.responseCount)
	for response := range request.responseChannel {
		responses = append(responses, response)
	}
	return responses
}

// GetResponseCount returns the number of responses received.
func (request *Request[T]) GetResponseCount() uint64 {
	return request.responseCount
}

// GetResponseLimit returns the response limit of the request.
func (request *Request[T]) GetResponseLimit() uint64 {
	return request.responseLimit
}

// Wait blocks until the request is done.
func (request *Request[T]) Wait() {
	<-request.doneChannel
}
