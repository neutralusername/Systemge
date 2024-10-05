package SyncRequest

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type SyncRequests struct {
	syncRequests    map[string]*syncRequest
	mutex           sync.Mutex
	randomizer      *Tools.Randomizer
	closeChannel    chan bool
	deadlineMs      uint64
	syncTokenLength uint32
}

type syncRequest struct {
	responseChannel chan *Message.Message
	abortChannel    chan bool
	responseLimit   uint64
}

func (syncRequests *SyncRequests) InitResponseChannel(responseLimit uint64) (string, <-chan *Message.Message) {
	syncRequests.mutex.Lock()
	defer syncRequests.mutex.Unlock()

	syncToken := syncRequests.randomizer.GenerateRandomString(syncRequests.syncTokenLength, Tools.ALPHA_NUMERIC)
	for _, ok := syncRequests.syncRequests[syncToken]; ok; {
		syncToken = syncRequests.randomizer.GenerateRandomString(syncRequests.syncTokenLength, Tools.ALPHA_NUMERIC)
	}
	syncRequestStruct := &syncRequest{
		responseChannel: make(chan *Message.Message, 1),
		abortChannel:    make(chan bool),
		responseLimit:   responseLimit,
	}
	syncRequests.syncRequests[syncToken] = syncRequestStruct

	var timeout <-chan time.Time
	if syncRequests.deadlineMs > 0 {
		timeout = time.After(time.Duration(syncRequests.deadlineMs) * time.Millisecond)
	}
	resChan := make(chan *Message.Message)
	go func() {
		select {
		case responseMessage := <-syncRequestStruct.responseChannel:
			resChan <- responseMessage
			close(resChan)

		case <-syncRequestStruct.abortChannel:
			close(resChan)

		case <-syncRequests.closeChannel:
			syncRequests.removeSyncRequest(syncToken)
			close(resChan)

		case <-timeout:
			syncRequests.removeSyncRequest(syncToken)
			close(resChan)
		}
	}()

	return syncToken, resChan
}

func (syncRequests *SyncRequests) AbortSyncRequest(syncToken string) error {
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

func (syncRequests *SyncRequests) AddSyncResponse(message *Message.Message) error {
	syncRequests.mutex.Lock()
	defer syncRequests.mutex.Unlock()

	syncRequestStruct, ok := syncRequests.syncRequests[message.GetSyncToken()]
	if !ok {
		return errors.New("no response channel found")
	}

	syncRequestStruct.responseChannel <- message
	close(syncRequestStruct.responseChannel)
	delete(syncRequests.syncRequests, message.GetSyncToken())

	return nil
}

func (syncRequests *SyncRequests) removeSyncRequest(syncToken string) error {
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
func (syncRequests *SyncRequests) GetOpenSyncRequests() []string {
	syncRequests.mutex.Lock()
	defer syncRequests.mutex.Unlock()
	syncTokens := make([]string, 0, len(syncRequests.syncRequests))
	for k := range syncRequests.syncRequests {
		syncTokens = append(syncTokens, k)
	}
	return syncTokens
}
