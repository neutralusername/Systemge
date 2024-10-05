package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Message"
)

type SyncManager struct {
	syncRequests    map[string]*SyncRequest
	mutex           sync.Mutex
	randomizer      *Randomizer
	syncTokenLength uint32
}

type SyncRequest struct {
	token           string
	responseChannel chan *Message.Message
	abortChannel    chan bool
	responseLimit   uint64
	responseCount   uint64
}

func (syncRequest *SyncRequest) GetToken() string {
	return syncRequest.token
}

func (syncRequest *SyncRequest) GetResponseChannel() <-chan *Message.Message {
	return syncRequest.responseChannel
}
func (syncRequest *SyncRequest) GetNextResponse() (*Message.Message, error) {
	response, ok := <-syncRequest.responseChannel
	if !ok {
		return nil, errors.New("response channel closed")
	}
	return response, nil
}

func NewSyncManager(syncTokenLength uint32, randomizerSeed int64) *SyncManager {
	return &SyncManager{
		syncRequests:    make(map[string]*SyncRequest),
		randomizer:      NewRandomizer(randomizerSeed),
		syncTokenLength: syncTokenLength,
	}
}

func (syncRequests *SyncManager) InitResponseChannel(responseLimit uint64, deadlineMs uint64) *SyncRequest {
	if responseLimit == 0 {
		responseLimit = 1
	}
	syncRequests.mutex.Lock()
	defer syncRequests.mutex.Unlock()

	syncToken := syncRequests.randomizer.GenerateRandomString(syncRequests.syncTokenLength, ALPHA_NUMERIC)
	for _, ok := syncRequests.syncRequests[syncToken]; ok; {
		syncToken = syncRequests.randomizer.GenerateRandomString(syncRequests.syncTokenLength, ALPHA_NUMERIC)
	}
	syncRequestStruct := &SyncRequest{
		token:           syncToken,
		responseChannel: make(chan *Message.Message, responseLimit),
		abortChannel:    make(chan bool),
		responseLimit:   responseLimit,
		responseCount:   0,
	}
	syncRequests.syncRequests[syncToken] = syncRequestStruct

	if deadlineMs > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(deadlineMs) * time.Millisecond):
				syncRequests.AbortSyncRequest(syncToken)
			case <-syncRequestStruct.abortChannel:
			}
		}()
	}

	return syncRequestStruct
}

func (syncRequests *SyncManager) AddSyncResponse(message *Message.Message) error {
	syncRequests.mutex.Lock()
	defer syncRequests.mutex.Unlock()

	syncRequestStruct, ok := syncRequests.syncRequests[message.GetSyncToken()]
	if !ok {
		return errors.New("no active sync request for token")
	}

	syncRequestStruct.responseChannel <- message
	syncRequestStruct.responseCount++

	if syncRequestStruct.responseCount >= syncRequestStruct.responseLimit {
		close(syncRequestStruct.responseChannel)
		delete(syncRequests.syncRequests, message.GetSyncToken())
	}

	return nil
}

func (syncRequests *SyncManager) AbortSyncRequest(syncToken string) error {
	syncRequests.mutex.Lock()
	defer syncRequests.mutex.Unlock()

	syncRequestStruct, ok := syncRequests.syncRequests[syncToken]
	if !ok {
		return errors.New("no active sync request for token")
	}

	close(syncRequestStruct.abortChannel)
	delete(syncRequests.syncRequests, syncToken)

	return nil
}

// returns a slice of syncTokens of open sync requests
func (syncRequests *SyncManager) GetOpenSyncRequests() []string {
	syncRequests.mutex.Lock()
	defer syncRequests.mutex.Unlock()
	syncTokens := make([]string, 0, len(syncRequests.syncRequests))
	for k := range syncRequests.syncRequests {
		syncTokens = append(syncTokens, k)
	}
	return syncTokens
}
