package Client

import (
	"Systemge/Message"
	"Systemge/Utilities"
	"time"
)

func (client *Client) handleBrokerMessages(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		messageBytes, err := brokerConnection.receive()
		if err != nil {
			brokerConnection.close()
			if client.IsStarted() {
				client.logger.Log(Utilities.NewError("Failed to receive message from message broker \""+brokerConnection.resolution.GetName()+"\"", err).Error())
			}

			client.mapOperationMutex.Lock()
			brokerConnection.mutex.Lock()
			delete(client.activeBrokerConnections, brokerConnection.resolution.GetAddress())
			subscribedTopics := make([]string, 0)
			for topic := range brokerConnection.topics {
				delete(client.topicResolutions, topic)
				if client.application.GetAsyncMessageHandlers()[topic] != nil || client.application.GetSyncMessageHandlers()[topic] != nil {
					subscribedTopics = append(subscribedTopics, topic)
				}
			}
			brokerConnection.topics = make(map[string]bool)
			brokerConnection.mutex.Unlock()
			client.mapOperationMutex.Unlock()

			if len(subscribedTopics) > 0 {
				for _, topic := range subscribedTopics {
					for client.IsStarted() {
						client.logger.Log("Attempting reconnect for topic \"" + topic + "\"")
						err := client.attemptToReconnectToSubscribedTopic(topic)
						if err == nil {
							client.logger.Log("Reconnect successful for topic \"" + topic + "\"")
							break
						}
						client.logger.Log(Utilities.NewError("Failed reconnect for topic \""+topic+"\"", err).Error())
						time.Sleep(1 * time.Second)
					}
				}
			}
			return
		}
		message := Message.Deserialize(messageBytes)
		if message == nil {
			client.logger.Log(Utilities.NewError("Failed to deserialize message \""+string(messageBytes)+"\"", nil).Error())
			continue
		}
		if message.GetSyncResponseToken() != "" {
			err := client.handleSyncResponse(message)
			if err != nil {
				client.logger.Log(Utilities.NewError("Failed to handle sync response", err).Error())
			}
		} else {
			err := client.handleMessage(message, brokerConnection)
			if err != nil {
				client.logger.Log(Utilities.NewError("Failed to handle message", err).Error())
			}
		}
	}
}

func (client *Client) handleSyncResponse(message *Message.Message) error {
	client.mapOperationMutex.Lock()
	responseChannel := client.messagesWaitingForResponse[message.GetSyncResponseToken()]
	client.mapOperationMutex.Unlock()
	if responseChannel == nil {
		return Utilities.NewError("No response channel for sync response token \""+message.GetSyncResponseToken()+"\"", nil)
	}
	responseChannel <- message
	return nil
}

func (client *Client) handleMessage(message *Message.Message, brokerConnection *brokerConnection) error {
	if !client.handleMessagesConcurrently {
		client.handleMessagesConcurrentlyMutex.Lock()
		defer client.handleMessagesConcurrentlyMutex.Unlock()
	}
	if message.GetSyncRequestToken() != "" {
		response, err := client.handleSyncMessage(message)
		if err != nil {
			errResponse := brokerConnection.send(message.NewResponse("error", client.name, Utilities.NewError("Error handling message", err).Error()))
			if errResponse != nil {
				return Utilities.NewError("Failed to send error response to message broker server", errResponse)
			}
			return Utilities.NewError("Error handling message", err)
		}
		err = brokerConnection.send(message.NewResponse(message.GetTopic(), client.name, response))
		if err != nil {
			return Utilities.NewError("Failed to send response to message broker server", err)
		}
	} else {
		err := client.handleAsyncMessage(message)
		if err != nil {
			return Utilities.NewError("Error handling message", err)
		}
	}
	return nil
}

func (client *Client) handleSyncMessage(message *Message.Message) (string, error) {
	client.mapOperationMutex.Lock()
	messageHandler := client.application.GetSyncMessageHandlers()[message.GetTopic()]
	client.mapOperationMutex.Unlock()
	if messageHandler == nil {
		return "", Utilities.NewError("No message handler for topic \""+message.GetTopic()+"\" on \""+client.name+"\"", nil)
	}
	response, err := messageHandler(message)
	if err != nil {
		return "", Utilities.NewError("Error handling message", err)
	} else {
		return response, nil
	}
}

func (client *Client) handleAsyncMessage(message *Message.Message) error {
	client.mapOperationMutex.Lock()
	messageHandler := client.application.GetAsyncMessageHandlers()[message.GetTopic()]
	client.mapOperationMutex.Unlock()
	if messageHandler == nil {
		return Utilities.NewError("No message handler for topic \""+message.GetTopic()+"\" on \""+client.name+"\"", nil)
	}
	err := messageHandler(message)
	if err != nil {
		return Utilities.NewError("Error handling message", err)
	}
	return nil
}
