package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"time"
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

func (broker *Broker) handleSyncRequest(nodeConnection *nodeConnection, message *Message.Message) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	if openSyncRequest := broker.openSyncRequests[message.GetSyncRequestToken()]; openSyncRequest != nil {
		return Error.New("token already in use", nil)
	}
	if len(broker.nodeSubscriptions[message.GetTopic()]) == 0 && message.GetTopic() != "subscribe" && message.GetTopic() != "unsubscribe" {
		return Error.New("no subscribers to topic \""+message.GetTopic()+"\"", nil)
	}
	syncRequest := newSyncRequest(nodeConnection, message)
	broker.openSyncRequests[message.GetSyncRequestToken()] = syncRequest
	go func() {
		timer := time.NewTimer(time.Duration(DEFAULT_SYNC_REQUEST_TIMEOUT) * time.Millisecond)
		defer timer.Stop()
		select {
		case response := <-syncRequest.responseChannel:
			broker.operationMutex.Lock()
			delete(broker.openSyncRequests, message.GetSyncRequestToken())
			broker.operationMutex.Unlock()
			err := nodeConnection.send(response)
			if err != nil {
				broker.logger.Log(Error.New("Failed to send response to node \""+nodeConnection.name+"\"", err).Error())
			}
		case <-timer.C:
			broker.operationMutex.Lock()
			delete(broker.openSyncRequests, message.GetSyncRequestToken())
			broker.operationMutex.Unlock()
			err := nodeConnection.send(message.NewResponse("error", broker.name, "request timed out"))
			if err != nil {
				broker.logger.Log(Error.New("Failed to send timeout response to node \""+nodeConnection.name+"\"", err).Error())
			}
		}
	}()
	return nil
}

func (broker *Broker) handleSyncResponse(message *Message.Message) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	waitingNodeConnection := broker.openSyncRequests[message.GetSyncResponseToken()]
	if waitingNodeConnection == nil {
		return Error.New("response to unknown sync request \""+message.GetSyncResponseToken()+"\"", nil)
	}
	waitingNodeConnection.responseChannel <- message
	return nil
}
