package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"strings"
)

func (server *Server) handleClientConnectionMessages(clientConnection *clientConnection) {
	for server.IsStarted() {
		messageBytes, err := clientConnection.receive()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") && !strings.Contains(err.Error(), "EOF") {
				server.logger.Log(Error.New("Failed to receive message from client \""+clientConnection.name+"\"", err).Error())
			}
			clientConnection.disconnect()
			return
		}
		message := Message.Deserialize(messageBytes)
		err = server.validateMessage(message)
		if err != nil {
			server.logger.Log(Error.New("Invalid message from client \""+clientConnection.name+"\"", err).Error())
			clientConnection.disconnect()
			return
		}
		err = server.handleMessage(clientConnection, message)
		if err != nil {
			server.logger.Log(Error.New("Failed to handle message from client \""+clientConnection.name+"\"", err).Error())
		}
	}
}

func (server *Server) validateMessage(message *Message.Message) error {
	if message == nil {
		return Error.New("Message is nil", nil)
	}
	if message.GetTopic() == "" {
		return Error.New("Message topic is empty", nil)
	}
	if message.GetOrigin() == "" {
		return Error.New("Message origin is empty", nil)
	}
	if message.GetSyncResponseToken() != "" && message.GetSyncRequestToken() != "" {
		return Error.New("Message cannot be both a sync request and a sync response", nil)
	}
	if message.GetSyncResponseToken() != "" {
		return nil
	}
	server.operationMutex.Lock()
	syncTopicExists := server.syncTopics[message.GetTopic()]
	asyncTopicExists := server.asyncTopics[message.GetTopic()]
	server.operationMutex.Unlock()
	if !syncTopicExists && !asyncTopicExists {
		return Error.New("Topic \""+message.GetTopic()+"\" does not exist on server \""+server.name+"\"", nil)
	}
	if syncTopicExists && message.GetSyncRequestToken() == "" {
		return Error.New("Topic \""+message.GetTopic()+"\" is a sync topic and message is not a sync request", nil)
	}
	if asyncTopicExists && message.GetSyncRequestToken() != "" {
		return Error.New("Topic \""+message.GetTopic()+"\" is an async topic and message is a sync request", nil)
	}
	return nil
}

func (server *Server) handleMessage(clientConnection *clientConnection, message *Message.Message) error {
	if message.GetSyncResponseToken() != "" {
		err := server.handleSyncResponse(message)
		if err != nil {
			server.logger.Log(Error.New("Failed to handle sync response from client \""+clientConnection.name+"\" with token \""+message.GetSyncResponseToken()+"\"", err).Error())
		}
		return nil
	}
	if message.GetSyncRequestToken() != "" {
		if err := server.handleSyncRequest(clientConnection, message); err != nil {
			//not using handleSyncResponse because the request failed, which means the syncRequest token has not been registered
			errResponse := clientConnection.send(message.NewResponse("error", server.name, Error.New("sync request failed", err).Error()))
			if errResponse != nil {
				server.logger.Log(Error.New("Failed to handle sync request from client \""+clientConnection.name+"\"", err).Error())
				return Error.New("failed to send error response", errResponse)
			}
			return Error.New("Failed to handle sync request from client \""+clientConnection.name+"\"", err)
		}
	}
	switch message.GetTopic() {
	case "heartbeat":
		err := clientConnection.resetWatchdog()
		if err != nil {
			return Error.New("Failed to reset watchdog for client \""+clientConnection.name+"\"", nil)
		}
		return nil
	case "unsubscribe":
		err := server.handleUnsubscribe(clientConnection, message)
		if err != nil {
			return Error.New("Failed to handle unsubscribe message from client \""+clientConnection.name+"\"", err)
		}
	case "subscribe":
		err := server.handleSubscribe(clientConnection, message)
		if err != nil {
			return Error.New("Failed to handle subscribe message from client \""+clientConnection.name+"\"", err)
		}
	case "consume":
		err := server.handleConsume(clientConnection)
		if err != nil {
			return Error.New("Failed to handle consume message from client \""+clientConnection.name+"\"", err)
		}
	default:
		server.propagateMessage(message)
	}
	return nil
}

func (server *Server) handleConsume(clientConnection *clientConnection) error {
	message := clientConnection.dequeueMessage_Timeout(DEFAULT_TCP_TIMEOUT)
	if message == nil {
		errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, "No message to consume"))
		if errResponse != nil {
			return Error.New("Failed to send error response to client \""+clientConnection.name+"\" and no message to consume", errResponse)
		}
		return Error.New("No message to consume for client \""+clientConnection.name+"\"", nil)
	}
	err := server.handleSyncResponse(message)
	if err != nil {
		return Error.New("Failed to send response to client \""+clientConnection.name+"\"", err)
	}
	return nil
}

func (server *Server) handleSubscribe(clientConnection *clientConnection, message *Message.Message) error {
	err := server.addSubscription(clientConnection, message.GetPayload())
	if err != nil {
		errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, Error.New("", err).Error()))
		if errResponse != nil {
			return Error.New("Failed to subscribe client \""+clientConnection.name+"\" to topic \""+message.GetPayload()+"\" and failed to send error response", errResponse)
		}
		return Error.New("Failed to subscribe client \""+clientConnection.name+"\" to topic \""+message.GetPayload()+"\"", err)
	}
	err = server.handleSyncResponse(message.NewResponse("subscribed", server.name, ""))
	if err != nil {
		return Error.New("Failed to send subscribe response to client \""+clientConnection.name+"\"", err)
	}
	return nil
}

func (server *Server) handleUnsubscribe(clientConnection *clientConnection, message *Message.Message) error {
	err := server.removeSubscription(clientConnection, message.GetPayload())
	if err != nil {
		errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, Error.New("", err).Error()))
		if errResponse != nil {
			return Error.New("Failed to unsubscribe client \""+clientConnection.name+"\" from topic \""+message.GetPayload()+"\" and failed to send error response", errResponse)
		}
		return Error.New("Failed to unsubscribe client \""+clientConnection.name+"\" from topic \""+message.GetPayload()+"\"", err)
	}
	err = server.handleSyncResponse(message.NewResponse("unsubscribed", server.name, ""))
	if err != nil {
		return Error.New("Failed to send unsubscribe response to client \""+clientConnection.name+"\"", err)
	}
	return nil
}

func (server *Server) propagateMessage(message *Message.Message) {
	server.operationMutex.Lock()
	clients := server.getSubscribers(message.GetTopic())
	server.operationMutex.Unlock()
	for _, clientConnection := range clients {
		if clientConnection.deliverImmediately {
			err := clientConnection.send(message)
			if err != nil {
				server.logger.Log(Error.New("Failed to send message to client \""+clientConnection.name+"\" with topic \""+message.GetTopic()+"\"", err).Error())
			}
		} else {
			clientConnection.queueMessage(message)
		}
	}
}
