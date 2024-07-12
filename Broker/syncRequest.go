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

func (broker *Broker) addSyncRequest(nodeConnection *nodeConnection, message *Message.Message) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	if openSyncRequest := broker.openSyncRequests[message.GetSyncRequestToken()]; openSyncRequest != nil {
		return Error.New("token is already in use", nil)
	}
	if len(broker.nodeSubscriptions[message.GetTopic()]) == 0 && message.GetTopic() != "subscribe" && message.GetTopic() != "unsubscribe" {
		return Error.New("no subscribers to topic \""+message.GetTopic()+"\"", nil)
	}
	syncRequest := newSyncRequest(nodeConnection, message)
	broker.openSyncRequests[message.GetSyncRequestToken()] = syncRequest
	go broker.handleSyncRequest(syncRequest)
	return nil
}

func (broker *Broker) handleSyncRequest(syncRequest *syncRequest) {
	timer := time.NewTimer(time.Duration(broker.config.SyncRequestTimeoutMs) * time.Millisecond)
	defer timer.Stop()
	select {
	case response := <-syncRequest.responseChannel:
		broker.operationMutex.Lock()
		delete(broker.openSyncRequests, syncRequest.message.GetSyncRequestToken())
		broker.operationMutex.Unlock()
		err := syncRequest.nodeConnection.send(response)
		if err != nil {
			broker.config.Logger.Warning(Error.New("failed to send sync response with topic \""+response.GetTopic()+"\" and token \""+response.GetSyncResponseToken()+"\" to node \""+syncRequest.nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", err).Error())
		} else {
			broker.config.Logger.Info(Error.New("sent sync response with topic \""+response.GetTopic()+"\" and token \""+response.GetSyncResponseToken()+"\" to node \""+syncRequest.nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", nil).Error())

		}
	case <-broker.stopChannel:
		broker.operationMutex.Lock()
		delete(broker.openSyncRequests, syncRequest.message.GetSyncRequestToken())
		broker.operationMutex.Unlock()
		err := syncRequest.nodeConnection.send(syncRequest.message.NewResponse("error", broker.GetName(), "broker stopped"))
		if err != nil {
			broker.config.Logger.Warning(Error.New("failed to send broker stopped sync response to node \""+syncRequest.nodeConnection.name+"\" with token \""+syncRequest.message.GetSyncRequestToken()+"\" on broker \""+broker.GetName()+"\"", err).Error())
		} else {
			broker.config.Logger.Info(Error.New("sent broker stopped sync response to node \""+syncRequest.nodeConnection.name+"\" with token \""+syncRequest.message.GetSyncRequestToken()+"\" on broker \""+broker.GetName()+"\"", nil).Error())
		}
	case <-timer.C:
		broker.operationMutex.Lock()
		delete(broker.openSyncRequests, syncRequest.message.GetSyncRequestToken())
		broker.operationMutex.Unlock()
		err := syncRequest.nodeConnection.send(syncRequest.message.NewResponse("error", broker.GetName(), "request timed out"))
		if err != nil {
			broker.config.Logger.Warning(Error.New("failed to send timeout sync response with token \""+syncRequest.message.GetSyncRequestToken()+"\" to node \""+syncRequest.nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", err).Error())
		} else {
			broker.config.Logger.Info(Error.New("sent timeout sync response with token \""+syncRequest.message.GetSyncRequestToken()+"\" to node \""+syncRequest.nodeConnection.name+"\" on broker \""+broker.GetName()+"\"", nil).Error())
		}
	}
}

func (broker *Broker) handleSyncResponse(message *Message.Message) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	waitingNodeConnection := broker.openSyncRequests[message.GetSyncResponseToken()]
	if waitingNodeConnection == nil {
		return Error.New("response to unknown sync request with token \""+message.GetSyncResponseToken()+"\"", nil)
	}
	waitingNodeConnection.responseChannel <- message
	return nil
}
