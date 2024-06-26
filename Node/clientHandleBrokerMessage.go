package Node

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (client *Node) handleBrokerMessages(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		messageBytes, err := brokerConnection.receive()
		if err != nil {
			brokerConnection.close()
			if client.IsStarted() {
				client.logger.Log(Error.New("Failed to receive message from message broker \""+brokerConnection.resolution.GetName()+"\"", err).Error())
			}
			client.handleBrokerDisconnect(brokerConnection)
			return
		}
		message := Message.Deserialize(messageBytes)
		if message == nil {
			client.logger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\"", nil).Error())
			continue
		}
		if message.GetSyncResponseToken() != "" {
			err := client.handleSyncResponse(message)
			if err != nil {
				client.logger.Log(Error.New("Failed to handle sync response", err).Error())
			}
		} else {
			err := client.handleMessage(message, brokerConnection)
			if err != nil {
				client.logger.Log(Error.New("Failed to handle message", err).Error())
			}
		}
	}
}

func (client *Node) handleSyncResponse(message *Message.Message) error {
	client.clientMutex.Lock()
	responseChannel := client.messagesWaitingForResponse[message.GetSyncResponseToken()]
	client.clientMutex.Unlock()
	if responseChannel == nil {
		return Error.New("No response channel for sync response token \""+message.GetSyncResponseToken()+"\"", nil)
	}
	responseChannel <- message
	return nil
}

func (client *Node) handleMessage(message *Message.Message, brokerConnection *brokerConnection) error {
	if client.config.HandleMessagesSequentially {
		client.handleMessagesSequentiallyMutex.Lock()
		defer client.handleMessagesSequentiallyMutex.Unlock()
	}
	if message.GetSyncRequestToken() != "" {
		response, err := client.handleSyncMessage(message)
		if err != nil {
			errResponse := brokerConnection.send(message.NewResponse("error", client.config.Name, Error.New("Error handling message", err).Error()))
			if errResponse != nil {
				return Error.New("Failed to send error response to message broker server", errResponse)
			}
			return Error.New("Error handling message", err)
		}
		err = brokerConnection.send(message.NewResponse(message.GetTopic(), client.config.Name, response))
		if err != nil {
			return Error.New("Failed to send response to message broker server", err)
		}
	} else {
		err := client.handleAsyncMessage(message)
		if err != nil {
			return Error.New("Error handling message", err)
		}
	}
	return nil
}

func (client *Node) handleSyncMessage(message *Message.Message) (string, error) {
	syncHandler := client.application.GetSyncMessageHandlers()[message.GetTopic()]
	if syncHandler == nil {
		return "", Error.New("No handler for topic \""+message.GetTopic()+"\"", nil)
	}
	response, err := syncHandler(client, message)
	if err != nil {
		return "", Error.New("Error handling message", err)
	}
	return response, nil
}

func (client *Node) handleAsyncMessage(message *Message.Message) error {
	asyncHandler := client.application.GetAsyncMessageHandlers()[message.GetTopic()]
	if asyncHandler == nil {
		return Error.New("No handler for topic \""+message.GetTopic()+"\"", nil)
	}
	err := asyncHandler(client, message)
	if err != nil {
		return Error.New("Error handling message", err)
	}
	return nil
}
