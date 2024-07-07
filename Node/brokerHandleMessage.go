package Node

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (node *Node) handleBrokerMessages(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		messageBytes, err := brokerConnection.receive()
		if err != nil {
			if node.IsStarted() {
				node.logger.Log(Error.New("Failed to receive message from message broker \""+brokerConnection.endpoint.GetAddress()+"\"", err).Error())
			}
			node.handleBrokerDisconnect(brokerConnection)
			return
		}
		message := Message.Deserialize(messageBytes)
		if message == nil {
			node.logger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\"", nil).Error())
			continue
		}
		if message.GetSyncResponseToken() != "" {
			err := node.handleSyncResponse(message)
			if err != nil {
				node.logger.Log(Error.New("Failed to handle sync response", err).Error())
			}
		} else {
			err := node.handleMessage(message, brokerConnection)
			if err != nil {
				node.logger.Log(Error.New("Failed to handle message", err).Error())
			}
		}
	}
}

func (node *Node) handleSyncResponse(message *Message.Message) error {
	node.mutex.Lock()
	responseChannel := node.messagesWaitingForResponse[message.GetSyncResponseToken()]
	node.mutex.Unlock()
	if responseChannel == nil {
		return Error.New("No response channel for sync response token \""+message.GetSyncResponseToken()+"\"", nil)
	}
	responseChannel <- message
	return nil
}

func (node *Node) handleMessage(message *Message.Message, brokerConnection *brokerConnection) error {
	if node.application.GetApplicationConfig().HandleMessagesSequentially {
		node.handleMessagesSequentiallyMutex.Lock()
		defer node.handleMessagesSequentiallyMutex.Unlock()
	}
	if message.GetSyncRequestToken() != "" {
		response, err := node.handleSyncMessage(message)
		if err != nil {
			errResponse := brokerConnection.send(message.NewResponse("error", node.config.Name, Error.New("Error handling message", err).Error()))
			if errResponse != nil {
				return Error.New("Failed to send error response to broker", errResponse)
			}
			return Error.New("Error handling message", err)
		}
		err = brokerConnection.send(message.NewResponse(message.GetTopic(), node.config.Name, response))
		if err != nil {
			return Error.New("Failed to send response to broker", err)
		}
	} else {
		err := node.handleAsyncMessage(message)
		if err != nil {
			return Error.New("Error handling message", err)
		}
	}
	return nil
}

func (node *Node) handleSyncMessage(message *Message.Message) (string, error) {
	syncHandler := node.application.GetSyncMessageHandlers()[message.GetTopic()]
	if syncHandler == nil {
		return "", Error.New("No handler for topic \""+message.GetTopic()+"\"", nil)
	}
	response, err := syncHandler(node, message)
	if err != nil {
		return "", Error.New("Error handling message", err)
	}
	return response, nil
}

func (node *Node) handleAsyncMessage(message *Message.Message) error {
	asyncHandler := node.application.GetAsyncMessageHandlers()[message.GetTopic()]
	if asyncHandler == nil {
		return Error.New("No handler for topic \""+message.GetTopic()+"\"", nil)
	}
	err := asyncHandler(node, message)
	if err != nil {
		return Error.New("Error handling message", err)
	}
	return nil
}
