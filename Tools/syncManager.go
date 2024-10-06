package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
)

type SyncManager struct {
	config   *Config.SyncManager
	requests map[string]*SyncRequest
	mutex    sync.Mutex
}

func NewSyncManager(config *Config.SyncManager) *SyncManager {
	return &SyncManager{
		requests: make(map[string]*SyncRequest),
		mutex:    sync.Mutex{},
		config:   config,
	}
}

func (manager *SyncManager) NewRequest(token string, responseLimit uint64, timeoutMs uint64) (*SyncRequest, error) {
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
	for _, ok := manager.requests[token]; ok; {
		return nil, errors.New("token already exists")
	}

	syncRequest := &SyncRequest{
		token:           token,
		responseChannel: make(chan any, responseLimit),
		abortChannel:    make(chan struct{}),
		responseLimit:   responseLimit,
		responseCount:   0,
	}
	manager.requests[token] = syncRequest

	if timeoutMs > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
				manager.AbortRequest(token)
			case <-syncRequest.abortChannel:
			}
		}()
	}

	return syncRequest, nil
}

func (manager *SyncManager) AddResponse(token string, response any) error {
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
		delete(manager.requests, token)
	}

	return nil
}

func (manager *SyncManager) AbortRequest(token string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	syncRequestStruct, ok := manager.requests[token]
	if !ok {
		return errors.New("no active sync request for token")
	}

	close(syncRequestStruct.abortChannel)
	delete(manager.requests, token)

	return nil
}

// returns a slice of syncTokens of active sync requests
func (manager *SyncManager) GetActiveRequests() []string {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	tokens := make([]string, 0, len(manager.requests))
	for k := range manager.requests {
		tokens = append(tokens, k)
	}
	return tokens
}

type SyncRequest struct {
	token           string
	responseChannel chan any
	abortChannel    chan struct{}
	responseLimit   uint64
	responseCount   uint64
}

func (request *SyncRequest) GetToken() string {
	return request.token
}

func (request *SyncRequest) GetResponseChannel() <-chan any {
	return request.responseChannel
}

func (request *SyncRequest) GetNextResponse() (any, error) {
	response, ok := <-request.responseChannel
	if !ok {
		return nil, errors.New("response channel closed")
	}
	return response, nil
}

func (request *SyncRequest) GetResponseCount() uint64 {
	return request.responseCount
}

func (request *SyncRequest) GetResponseLimit() uint64 {
	return request.responseLimit
}
