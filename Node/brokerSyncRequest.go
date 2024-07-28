package Node

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

type syncRequest struct {
	nodeConnection  *nodeConnection
	message         *Message.Message
	responseChannel chan *Message.Message
}

func newSyncRequest(nodeConnection *nodeConnection, message *Message.Message) *syncRequest {
	return &syncRequest{
		nodeConnection:  nodeConnection,
		message:         message,
		responseChannel: make(chan *Message.Message, 1),
	}
}

func (broker *brokerComponent) addSyncRequest(nodeConnection *nodeConnection, message *Message.Message) (*syncRequest, error) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()
	if openSyncRequest := broker.openSyncRequests[message.GetSyncRequestToken()]; openSyncRequest != nil {
		return nil, Error.New("token is already in use", nil)
	}
	if len(broker.nodeSubscriptions[message.GetTopic()]) == 0 && message.GetTopic() != "subscribe" && message.GetTopic() != "unsubscribe" {
		return nil, Error.New("no subscribers to topic \""+message.GetTopic()+"\"", nil)
	}
	syncRequest := newSyncRequest(nodeConnection, message)
	broker.openSyncRequests[message.GetSyncRequestToken()] = syncRequest
	return syncRequest, nil
}

func (broker *brokerComponent) handleSyncRequest(syncRequest *syncRequest) (*Message.Message, error) {
	timer := time.NewTimer(time.Duration(broker.application.GetBrokerComponentConfig().SyncResponseTimeoutMs) * time.Millisecond)
	defer timer.Stop()
	select {
	case response := <-syncRequest.responseChannel:
		if response == nil {
			return nil, Error.New("broker component stopped before sending response to sync request with token \""+syncRequest.message.GetSyncRequestToken()+"\"", nil)
		}
		return response, nil
	case <-timer.C:
		return nil, Error.New("sync request with token \""+syncRequest.message.GetSyncRequestToken()+"\" timed out", nil)
	}
}

func (broker *brokerComponent) handleSyncResponse(message *Message.Message) error {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()
	syncRequest := broker.openSyncRequests[message.GetSyncResponseToken()]
	if syncRequest == nil {
		return Error.New("response to unknown sync request with token \""+message.GetSyncResponseToken()+"\"", nil)
	}
	delete(broker.openSyncRequests, message.GetSyncResponseToken())
	syncRequest.responseChannel <- message
	return nil
}
