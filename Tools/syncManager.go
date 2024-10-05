package Tools

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Message"
)

type SyncManager struct {
	syncRequests    map[string]*SyncRequest
	mutex           sync.Mutex
	randomizer      *Randomizer
	closeChannel    chan bool
	deadlineMs      uint64
	syncTokenLength uint32
}

type SyncRequest struct {
	responseChannel chan *Message.Message
	abortChannel    chan bool
	responseLimit   uint64
	responseCount   uint64
}

func (syncRequests *SyncManager) InitResponseChannel(responseLimit uint64) (string, *SyncRequest) {
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
		responseChannel: make(chan *Message.Message, responseLimit),
		abortChannel:    make(chan bool),
		responseLimit:   responseLimit,
	}
	syncRequests.syncRequests[syncToken] = syncRequestStruct

	return syncToken, syncRequestStruct
}

func (syncRequests *SyncManager) AddSyncResponse(message *Message.Message) error {
	syncRequests.mutex.Lock()
	defer syncRequests.mutex.Unlock()

	syncRequestStruct, ok := syncRequests.syncRequests[message.GetSyncToken()]
	if !ok {
		return errors.New("no response channel found")
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
		return errors.New("no response channel found")
	}

	close(syncRequestStruct.abortChannel)
	delete(syncRequests.syncRequests, syncToken)

	return nil
}

func (syncRequests *SyncManager) removeSyncRequest(syncToken string) error {
	syncRequests.mutex.Lock()
	defer syncRequests.mutex.Unlock()

	_, ok := syncRequests.syncRequests[syncToken]
	if !ok {
		return errors.New("no response channel found")
	}
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
