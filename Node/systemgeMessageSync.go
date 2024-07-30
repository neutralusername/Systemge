package Node

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type SyncResponseChannel struct {
	closeChannel   chan struct{}
	channel        chan *SyncResponse
	requestMessage *Message.Message
}

func (syncResponseChannel *SyncResponseChannel) Close() {
	close(syncResponseChannel.closeChannel)
}

type SyncResponse struct {
	responseMessage *Message.Message
	origin          string
}

func newSyncResponseChannel(requestMessage *Message.Message) *SyncResponseChannel {
	return &SyncResponseChannel{
		closeChannel:   make(chan struct{}),
		channel:        make(chan *SyncResponse),
		requestMessage: requestMessage,
	}
}

func (syncResponse *SyncResponse) GetMessage() *Message.Message {
	return syncResponse.responseMessage
}

func (syncResponse *SyncResponse) GetOrigin() string {
	return syncResponse.origin
}

// blocks until response is received
func (syncResponseChannel *SyncResponseChannel) ReceiveResponse() (*SyncResponse, error) {
	syncResponse := <-syncResponseChannel.channel
	if syncResponse == nil {
		return nil, Error.New("Channel closed", nil)
	}
	return syncResponse, nil
}

func (syncResponseChannel *SyncResponseChannel) GetRequestMessage() *Message.Message {
	return syncResponseChannel.requestMessage
}

// blocks until response is received or timeout is reached
func (syncResponseChannel *SyncResponseChannel) ReceiveResponseTimeout(timeoutMs uint64) (*SyncResponse, error) {
	timeout := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	select {
	case syncResponse := <-syncResponseChannel.channel:
		timeout.Stop()
		if syncResponse == nil {
			return nil, Error.New("Channel closed", nil)
		}
		return syncResponse, nil
	case <-timeout.C:
		return nil, Error.New("Timeout", nil)
	}
}

func (node *Node) SyncMessage(topic, payload string) (*SyncResponseChannel, error) {
	if systemge := node.systemge; systemge != nil {
		message := Message.NewSync(topic, payload, node.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC))
		responseChannel := newSyncResponseChannel(message)
		systemge.syncRequestMutex.Lock()
		systemge.syncRequestChannels[message.GetSyncTokenToken()] = responseChannel
		systemge.syncRequestMutex.Unlock()
		for _, outgoingConnection := range systemge.topicResolutions[topic] {
			err := systemge.sendOutgoingConnection(outgoingConnection, message)
			if err != nil {
				if errorLogger := node.GetErrorLogger(); errorLogger != nil {
					errorLogger.Log(Error.New("Failed to send sync message with topic \""+topic+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
			}
		}
		go func() {
			if syncRequestTimeoutMs := systemge.config.SyncRequestTimeoutMs; syncRequestTimeoutMs > 0 {
				timeout := time.NewTimer(time.Duration(syncRequestTimeoutMs) * time.Millisecond)
				select {
				case <-timeout.C:
				case <-node.stopChannel:
				case <-responseChannel.closeChannel:
				}
				timeout.Stop()
			} else {
				select {
				case <-node.stopChannel:
				case <-responseChannel.closeChannel:
				}
			}
			systemge.syncRequestMutex.Lock()
			close(responseChannel.channel)
			delete(systemge.syncRequestChannels, message.GetSyncTokenToken())
			systemge.syncRequestMutex.Unlock()
		}()
		return responseChannel, nil
	}
	return nil, Error.New("systemge not initialized", nil)
}
