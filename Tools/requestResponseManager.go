package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
)

type RequestResponseManager[T any] struct {
	config   *Config.SyncManager
	requests map[string]*request[T]
	mutex    sync.Mutex
}
type request[T any] struct {
	token           string
	responseChannel chan T
	doneChannel     chan struct{}
	responseLimit   uint64
	responseCount   uint64
}

func NewSyncManager[T any](config *Config.SyncManager) *RequestResponseManager[T] {
	return &RequestResponseManager[T]{
		requests: make(map[string]*request[T]),
		mutex:    sync.Mutex{},
		config:   config,
	}
}

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

	syncRequest := &request[T]{
		token:           token,
		responseChannel: make(chan T, responseLimit),
		doneChannel:     make(chan struct{}),
		responseLimit:   responseLimit,
		responseCount:   0,
	}
	manager.requests[token] = syncRequest

	if timeoutMs > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
				manager.AbortRequest(token)
			case <-syncRequest.doneChannel:
			}
		}()
	}

	return syncRequest, nil
}

func (manager *RequestResponseManager[T]) AddResponse(token string, response T) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	syncRequest, ok := manager.requests[token]
	if !ok {
		return errors.New("no active sync request for token")
	}

	syncRequest.responseChannel <- response
	syncRequest.responseCount++

	if syncRequest.responseCount >= syncRequest.responseLimit {
		close(syncRequest.responseChannel)
		close(syncRequest.doneChannel)
		delete(manager.requests, token)
	}

	return nil
}

func (manager *RequestResponseManager[T]) AbortRequest(token string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	syncRequestStruct, ok := manager.requests[token]
	if !ok {
		return errors.New("no active sync request for token")
	}

	close(syncRequestStruct.doneChannel)
	close(syncRequestStruct.responseChannel)
	delete(manager.requests, token)

	return nil
}

// returns a slice of syncTokens of active sync requests
func (manager *RequestResponseManager[T]) GetActiveRequests() []string {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	tokens := make([]string, 0, len(manager.requests))
	for k := range manager.requests {
		tokens = append(tokens, k)
	}
	return tokens
}

func (request *request[T]) GetToken() string {
	return request.token
}

func (request *request[T]) GetResponseChannel() <-chan T {
	return request.responseChannel
}

func (request *request[T]) GetNextResponse() (T, error) {
	response, ok := <-request.responseChannel
	if !ok {
		var nilValue T
		return nilValue, errors.New("response channel closed")
	}
	return response, nil
}

func (request *request[T]) GetResponseCount() uint64 {
	return request.responseCount
}

func (request *request[T]) GetResponseLimit() uint64 {
	return request.responseLimit
}
