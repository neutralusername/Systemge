package Broker

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"strings"
)

func (server *Server) handleBrokerClientMessages(clientConnection *clientConnection) {
	for server.IsStarted() {
		messageBytes, err := clientConnection.receive()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") && !strings.Contains(err.Error(), "EOF") {
				server.logger.Log(Utilities.NewError("Failed to receive message from client \""+clientConnection.name+"\"", err).Error())
			}
			clientConnection.disconnect()
			return
		}
		message := Message.Deserialize(messageBytes)
		err = server.validateMessage(message)
		if err != nil {
			server.logger.Log(Utilities.NewError("Invalid message from client \""+clientConnection.name+"\"", nil).Error())
			clientConnection.disconnect()
			return
		}
		err = server.handleMessage(clientConnection, message)
		if err != nil {
			server.logger.Log(Utilities.NewError("Failed to handle message from client \""+clientConnection.name+"\"", err).Error())
		}
	}
}

func (server *Server) validateMessage(message *Message.Message) error {
	if message == nil {
		return Utilities.NewError("Message is nil", nil)
	}
	if message.GetTopic() == "" {
		return Utilities.NewError("Message topic is empty", nil)
	}
	if message.GetOrigin() == "" {
		return Utilities.NewError("Message origin is empty", nil)
	}
	if message.GetSyncResponseToken() != "" && message.GetSyncRequestToken() != "" {
		return Utilities.NewError("Message cannot be both a sync request and a sync response", nil)
	}
	if message.GetSyncResponseToken() != "" {
		return nil
	}
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if !server.syncTopics[message.GetTopic()] && !server.asyncTopics[message.GetTopic()] {
		return Utilities.NewError("Topic \""+message.GetTopic()+"\" does not exist on server \""+server.name+"\"", nil)
	}
	return nil
}

func (server *Server) handleMessage(clientConnection *clientConnection, message *Message.Message) error {
	if message.GetSyncResponseToken() != "" {
		err := server.handleSyncResponse(message)
		if err != nil {
			server.logger.Log(Utilities.NewError("Failed to handle sync response from client \""+clientConnection.name+"\"", err).Error())
		}
		return nil
	}
	if message.GetSyncRequestToken() != "" {
		if err := server.handleSyncRequest(clientConnection, message); err != nil {
			errReponse := server.handleSyncResponse(message.NewResponse("error", server.name, Utilities.NewError("sync request failed", err).Error()))
			if errReponse != nil {
				return Utilities.NewError("Failed to handle sync request from client \""+clientConnection.name+"\" and failed to send error response", errReponse)
			}
			return Utilities.NewError("Failed to handle sync request from client \""+clientConnection.name+"\"", err)
		}
	}
	switch message.GetTopic() {
	case "heartbeat":
		err := clientConnection.resetWatchdog()
		if err != nil {
			return Utilities.NewError("Failed to reset watchdog for client \""+clientConnection.name+"\"", nil)
		}
		return nil
	case "unsubscribe":
		err := server.handleUnsubscribe(clientConnection, message)
		if err != nil {
			return Utilities.NewError("Failed to handle unsubscribe message from client \""+clientConnection.name+"\"", err)
		}
	case "subscribe":
		err := server.handleSubscribe(clientConnection, message)
		if err != nil {
			return Utilities.NewError("Failed to handle subscribe message from client \""+clientConnection.name+"\"", err)
		}
	case "consume":
		err := server.handleConsume(clientConnection)
		if err != nil {
			return Utilities.NewError("Failed to handle consume message from client \""+clientConnection.name+"\"", err)
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
			return Utilities.NewError("Failed to send error response to client \""+clientConnection.name+"\" and no message to consume", errResponse)
		}
		return Utilities.NewError("No message to consume for client \""+clientConnection.name+"\"", nil)
	}
	err := server.handleSyncResponse(message)
	if err != nil {
		return Utilities.NewError("Failed to send response to client \""+clientConnection.name+"\"", err)
	}
	return nil
}

func (server *Server) handleSubscribe(clientConnection *clientConnection, message *Message.Message) error {
	err := server.addSubscription(clientConnection, message.GetPayload())
	if err != nil {
		errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, Utilities.NewError("", err).Error()))
		if errResponse != nil {
			return Utilities.NewError("Failed to subscribe client \""+clientConnection.name+"\" to topic \""+message.GetPayload()+"\" and failed to send error response", errResponse)
		}
		return Utilities.NewError("Failed to subscribe client \""+clientConnection.name+"\" to topic \""+message.GetPayload()+"\"", err)
	}
	err = server.handleSyncResponse(message.NewResponse("subscribed", server.name, ""))
	if err != nil {
		return Utilities.NewError("Failed to send subscribe response to client \""+clientConnection.name+"\"", err)
	}
	return nil
}

func (server *Server) handleUnsubscribe(clientConnection *clientConnection, message *Message.Message) error {
	err := server.removeSubscription(clientConnection, message.GetPayload())
	if err != nil {
		errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, Utilities.NewError("", err).Error()))
		if errResponse != nil {
			return Utilities.NewError("Failed to unsubscribe client \""+clientConnection.name+"\" from topic \""+message.GetPayload()+"\" and failed to send error response", errResponse)
		}
		return Utilities.NewError("Failed to unsubscribe client \""+clientConnection.name+"\" from topic \""+message.GetPayload()+"\"", err)
	}
	err = server.handleSyncResponse(message.NewResponse("unsubscribed", server.name, ""))
	if err != nil {
		return Utilities.NewError("Failed to send unsubscribe response to client \""+clientConnection.name+"\"", err)
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
				server.logger.Log(Utilities.NewError("Failed to send message to client \""+clientConnection.name+"\" with topic \""+message.GetTopic()+"\"", err).Error())
			}
		} else {
			clientConnection.queueMessage(message)
		}
	}
}
