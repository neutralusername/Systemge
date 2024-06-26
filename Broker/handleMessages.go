package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"strings"
)

func (server *Server) handleNodeConnectionMessages(nodeConnection *nodeConnection) {
	for server.IsStarted() {
		messageBytes, err := nodeConnection.receive()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") && !strings.Contains(err.Error(), "EOF") {
				server.logger.Log(Error.New("Failed to receive message from node \""+nodeConnection.name+"\"", err).Error())
			}
			nodeConnection.disconnect()
			return
		}
		message := Message.Deserialize(messageBytes)
		err = server.validateMessage(message)
		if err != nil {
			server.logger.Log(Error.New("Invalid message from node \""+nodeConnection.name+"\"", err).Error())
			nodeConnection.disconnect()
			return
		}
		err = server.handleMessage(nodeConnection, message)
		if err != nil {
			server.logger.Log(Error.New("Failed to handle message from node \""+nodeConnection.name+"\"", err).Error())
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

func (server *Server) handleMessage(nodeConnection *nodeConnection, message *Message.Message) error {
	if message.GetSyncResponseToken() != "" {
		err := server.handleSyncResponse(message)
		if err != nil {
			server.logger.Log(Error.New("Failed to handle sync response from node \""+nodeConnection.name+"\" with token \""+message.GetSyncResponseToken()+"\"", err).Error())
		}
		return nil
	}
	if message.GetSyncRequestToken() != "" {
		if err := server.handleSyncRequest(nodeConnection, message); err != nil {
			//not using handleSyncResponse because the request failed, which means the syncRequest token has not been registered
			errResponse := nodeConnection.send(message.NewResponse("error", server.name, Error.New("sync request failed", err).Error()))
			if errResponse != nil {
				server.logger.Log(Error.New("Failed to handle sync request from node \""+nodeConnection.name+"\"", err).Error())
				return Error.New("failed to send error response", errResponse)
			}
			return Error.New("Failed to handle sync request from node \""+nodeConnection.name+"\"", err)
		}
	}
	switch message.GetTopic() {
	case "heartbeat":
		err := nodeConnection.resetWatchdog()
		if err != nil {
			return Error.New("Failed to reset watchdog for node \""+nodeConnection.name+"\"", nil)
		}
		return nil
	case "unsubscribe":
		err := server.handleUnsubscribe(nodeConnection, message)
		if err != nil {
			return Error.New("Failed to handle unsubscribe message from node \""+nodeConnection.name+"\"", err)
		}
	case "subscribe":
		err := server.handleSubscribe(nodeConnection, message)
		if err != nil {
			return Error.New("Failed to handle subscribe message from node \""+nodeConnection.name+"\"", err)
		}
	case "consume":
		err := server.handleConsume(nodeConnection)
		if err != nil {
			return Error.New("Failed to handle consume message from node \""+nodeConnection.name+"\"", err)
		}
	default:
		server.propagateMessage(message)
	}
	return nil
}

func (server *Server) handleConsume(nodeConnection *nodeConnection) error {
	message := nodeConnection.dequeueMessage_Timeout(DEFAULT_TCP_TIMEOUT)
	if message == nil {
		errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, "No message to consume"))
		if errResponse != nil {
			return Error.New("Failed to send error response to node \""+nodeConnection.name+"\" and no message to consume", errResponse)
		}
		return Error.New("No message to consume for node \""+nodeConnection.name+"\"", nil)
	}
	err := server.handleSyncResponse(message)
	if err != nil {
		return Error.New("Failed to send response to node \""+nodeConnection.name+"\"", err)
	}
	return nil
}

func (server *Server) handleSubscribe(nodeConnection *nodeConnection, message *Message.Message) error {
	err := server.addSubscription(nodeConnection, message.GetPayload())
	if err != nil {
		errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, Error.New("", err).Error()))
		if errResponse != nil {
			return Error.New("Failed to subscribe node \""+nodeConnection.name+"\" to topic \""+message.GetPayload()+"\" and failed to send error response", errResponse)
		}
		return Error.New("Failed to subscribe node \""+nodeConnection.name+"\" to topic \""+message.GetPayload()+"\"", err)
	}
	err = server.handleSyncResponse(message.NewResponse("subscribed", server.name, ""))
	if err != nil {
		return Error.New("Failed to send subscribe response to node \""+nodeConnection.name+"\"", err)
	}
	return nil
}

func (server *Server) handleUnsubscribe(nodeConnection *nodeConnection, message *Message.Message) error {
	err := server.removeSubscription(nodeConnection, message.GetPayload())
	if err != nil {
		errResponse := server.handleSyncResponse(message.NewResponse("error", server.name, Error.New("", err).Error()))
		if errResponse != nil {
			return Error.New("Failed to unsubscribe node \""+nodeConnection.name+"\" from topic \""+message.GetPayload()+"\" and failed to send error response", errResponse)
		}
		return Error.New("Failed to unsubscribe node \""+nodeConnection.name+"\" from topic \""+message.GetPayload()+"\"", err)
	}
	err = server.handleSyncResponse(message.NewResponse("unsubscribed", server.name, ""))
	if err != nil {
		return Error.New("Failed to send unsubscribe response to node \""+nodeConnection.name+"\"", err)
	}
	return nil
}

func (server *Server) propagateMessage(message *Message.Message) {
	server.operationMutex.Lock()
	nodes := server.getSubscribedNodes(message.GetTopic())
	server.operationMutex.Unlock()
	for _, nodeConnection := range nodes {
		if nodeConnection.deliverImmediately {
			err := nodeConnection.send(message)
			if err != nil {
				server.logger.Log(Error.New("Failed to send message to node \""+nodeConnection.name+"\" with topic \""+message.GetTopic()+"\"", err).Error())
			}
		} else {
			nodeConnection.queueMessage(message)
		}
	}
}
