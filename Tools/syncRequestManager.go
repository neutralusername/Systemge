package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Message"
)

type SyncRequestManager struct {
	requests    map[string]*SyncRequest
	mutex       sync.Mutex
	randomizer  *Randomizer
	tokenLength uint32
}

type SyncRequest struct {
	token           string
	responseChannel chan *Message.Message
	abortChannel    chan bool
	responseLimit   uint64
	responseCount   uint64
}

func (request *SyncRequest) GetToken() string {
	return request.token
}

func (request *SyncRequest) GetResponseChannel() <-chan *Message.Message {
	return request.responseChannel
}
func (request *SyncRequest) GetNextResponse() (*Message.Message, error) {
	response, ok := <-request.responseChannel
	if !ok {
		return nil, errors.New("response channel closed")
	}
	return response, nil
}

func NewSyncRequestManager(tokenLength uint32, randomizerSeed int64) *SyncRequestManager {
	return &SyncRequestManager{
		requests:    make(map[string]*SyncRequest),
		randomizer:  NewRandomizer(randomizerSeed),
		tokenLength: tokenLength,
	}
}

func (manager *SyncRequestManager) NewRequest(responseLimit uint64, deadlineMs uint64) *SyncRequest {
	if responseLimit == 0 {
		responseLimit = 1
	}
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	token := manager.randomizer.GenerateRandomString(manager.tokenLength, ALPHA_NUMERIC)
	for _, ok := manager.requests[token]; ok; {
		token = manager.randomizer.GenerateRandomString(manager.tokenLength, ALPHA_NUMERIC)
	}
	syncRequest := &SyncRequest{
		token:           token,
		responseChannel: make(chan *Message.Message, responseLimit),
		abortChannel:    make(chan bool),
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

func (manager *SyncRequestManager) AddResponse(message *Message.Message) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	syncRequest, ok := manager.requests[message.GetSyncToken()]
	if !ok {
		return errors.New("no active sync request for token")
	}

	syncRequest.responseChannel <- message
	syncRequest.responseCount++

	if syncRequest.responseCount >= syncRequest.responseLimit {
		close(syncRequest.responseChannel)
		delete(manager.requests, message.GetSyncToken())
	}

	return nil
}

func (manager *SyncRequestManager) AbortRequest(token string) error {
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
func (manager *SyncRequestManager) GetActiveRequests() []string {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	tokens := make([]string, 0, len(manager.requests))
	for k := range manager.requests {
		tokens = append(tokens, k)
	}
	return tokens
}
