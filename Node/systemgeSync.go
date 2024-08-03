package Node

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

type SyncResponseChannel struct {
	closeChannel    chan struct{}
	responseChannel chan *Message.Message
	responseCount   uint64
	requestMessage  *Message.Message
}

func newSyncResponseChannel(requestMessage *Message.Message, responseLimit uint64) *SyncResponseChannel {
	return &SyncResponseChannel{
		closeChannel:    make(chan struct{}),
		responseChannel: make(chan *Message.Message, responseLimit),
		requestMessage:  requestMessage,
		responseCount:   0,
	}
}

func (syncResponseChannel *SyncResponseChannel) Close() {
	close(syncResponseChannel.closeChannel)
}

func (systemge *systemgeComponent) handleSyncResponse(message *Message.Message) error {
	systemge.syncRequestMutex.Lock()
	defer systemge.syncRequestMutex.Unlock()
	syncResponseChannel := systemge.syncRequestChannels[message.GetSyncTokenToken()]
	if syncResponseChannel == nil {
		return Error.New("Received sync response for unknown token", nil)
	}
	if syncResponseChannel.responseCount >= systemge.config.SyncResponseLimit {
		return Error.New("Sync response limit reached", nil)
	}
	if message.GetTopic() == Message.TOPIC_SUCCESS {
		systemge.incomingSyncSuccessResponses.Add(1)
	} else {
		systemge.incomingSyncFailureResponses.Add(1)
	}
	syncResponseChannel.responseCount++
	systemge.incomingSyncResponses.Add(1)
	syncResponseChannel.responseChannel <- message
	return nil
}

// blocks until response is received
func (syncResponseChannel *SyncResponseChannel) ReceiveResponse() (*Message.Message, error) {
	syncResponse := <-syncResponseChannel.responseChannel
	if syncResponse == nil {
		return nil, Error.New("Channel closed", nil)
	}
	return syncResponse, nil
}

func (syncResponseChannel *SyncResponseChannel) GetRequestMessage() *Message.Message {
	return syncResponseChannel.requestMessage
}

// blocks until response is received or timeout is reached
func (syncResponseChannel *SyncResponseChannel) ReceiveResponseTimeout(timeoutMs uint64) (*Message.Message, error) {
	timeout := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	select {
	case syncResponse := <-syncResponseChannel.responseChannel:
		timeout.Stop()
		if syncResponse == nil {
			return nil, Error.New("Channel closed", nil)
		}
		return syncResponse, nil
	case <-timeout.C:
		return nil, Error.New("Timeout", nil)
	}
}
