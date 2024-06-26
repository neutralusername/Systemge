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

func (server *Server) handleSyncRequest(nodeConnection *nodeConnection, message *Message.Message) error {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if openSyncRequest := server.openSyncRequests[message.GetSyncRequestToken()]; openSyncRequest != nil {
		return Error.New("token already in use", nil)
	}
	if len(server.nodeSubscriptions[message.GetTopic()]) == 0 && message.GetTopic() != "subscribe" && message.GetTopic() != "unsubscribe" {
		return Error.New("no subscribers to topic \""+message.GetTopic()+"\"", nil)
	}
	syncRequest := newSyncRequest(nodeConnection, message)
	server.openSyncRequests[message.GetSyncRequestToken()] = syncRequest
	go func() {
		timer := time.NewTimer(time.Duration(DEFAULT_SYNC_REQUEST_TIMEOUT) * time.Millisecond)
		defer timer.Stop()
		select {
		case response := <-syncRequest.responseChannel:
			server.operationMutex.Lock()
			delete(server.openSyncRequests, message.GetSyncRequestToken())
			server.operationMutex.Unlock()
			err := nodeConnection.send(response)
			if err != nil {
				server.logger.Log(Error.New("Failed to send response to node \""+nodeConnection.name+"\"", err).Error())
			}
		case <-timer.C:
			server.operationMutex.Lock()
			delete(server.openSyncRequests, message.GetSyncRequestToken())
			server.operationMutex.Unlock()
			err := nodeConnection.send(message.NewResponse("error", server.name, "request timed out"))
			if err != nil {
				server.logger.Log(Error.New("Failed to send timeout response to node \""+nodeConnection.name+"\"", err).Error())
			}
		}
	}()
	return nil
}

func (server *Server) handleSyncResponse(message *Message.Message) error {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	waitingNodeConnection := server.openSyncRequests[message.GetSyncResponseToken()]
	if waitingNodeConnection == nil {
		return Error.New("response to unknown sync request \""+message.GetSyncResponseToken()+"\"", nil)
	}
	waitingNodeConnection.responseChannel <- message
	return nil
}
