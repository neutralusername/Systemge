package Tools

import (
	"errors"
	"math"
	"sync"
	"time"
)

type SyncManager struct {
	requests       map[string]*SyncRequest
	mutex          sync.Mutex
	randomizer     *Randomizer
	tokenLength    uint32
	maxTotalTokens float64
}

func NewSyncManager(tokenLength uint32, tokenAlphabet string, randomizerSeed int64) *SyncManager {
	return &SyncManager{
		requests:       make(map[string]*SyncRequest),
		randomizer:     NewRandomizer(randomizerSeed),
		tokenLength:    tokenLength,
		maxTotalTokens: math.Pow(float64(len(tokenAlphabet)), float64(tokenLength)) * 0.9,
	}
}

func (manager *SyncManager) NewRequest(responseLimit uint64, deadlineMs uint64) (*SyncRequest, error) {
	if responseLimit == 0 {
		responseLimit = 1
	}
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if len(manager.requests) >= int(manager.maxTotalTokens) {
		return nil, errors.New("maximum number of active sync requests reached")
	}

	token := manager.randomizer.GenerateRandomString(manager.tokenLength, ALPHA_NUMERIC)
	for _, ok := manager.requests[token]; ok; {
		token = manager.randomizer.GenerateRandomString(manager.tokenLength, ALPHA_NUMERIC)
	}
	syncRequest := &SyncRequest{
		token:           token,
		responseChannel: make(chan any, responseLimit),
		abortChannel:    make(chan struct{}),
		responseLimit:   responseLimit,
		responseCount:   0,
	}
	manager.requests[token] = syncRequest

	if deadlineMs > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(deadlineMs) * time.Millisecond):
				manager.AbortRequest(token)
			case <-syncRequest.abortChannel:
			}
		}()
	}

	return syncRequest
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
