package Client

import (
	"Systemge/Error"
	"Systemge/Message"
	"strings"
)

func (client *Client) handleServerMessages(serverConnection *brokerConnection) {
	for serverConnection.netConn != nil {
		messageBytes, err := serverConnection.receive()
		if err != nil {
			serverConnection.close()
			if !strings.Contains(err.Error(), "use of closed network connection") { // do not attempt to reconnect if the connection was closed from the client side
				client.attemptToReconnect(serverConnection)
			}
			return
		}
		message := Message.Deserialize(messageBytes)
		if message == nil {
			client.logger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\"", nil).Error())
			continue
		}
		if message.GetSyncResponseToken() != "" {
			client.mapOperationMutex.Lock()
			responseChannel := client.messagesWaitingForResponse[message.GetSyncResponseToken()]
			client.mapOperationMutex.Unlock()
			if responseChannel != nil {
				responseChannel <- message
			} else {
				client.logger.Log(Error.New("No response channel for sync response token \""+message.GetSyncResponseToken()+"\"", nil).Error())
			}
		} else {
			client.handleMessage(message, serverConnection)
		}
	}
}

func (client *Client) handleMessage(message *Message.Message, serverConnection *brokerConnection) {
	if !client.handleServerMessagesConcurrently {
		client.handleServerMessagesConcurrentlyMutex.Lock()
	}
	if message.GetSyncRequestToken() != "" {
		response, err := client.handleSyncMessage(message)
		if !client.handleServerMessagesConcurrently {
			client.handleServerMessagesConcurrentlyMutex.Unlock()
		}
		if err != nil {
			err := serverConnection.send(message.NewResponse("error", client.name, Error.New("Error handling message", err).Error()))
			if err != nil {
				client.logger.Log(Error.New("Failed to send error response to message broker server", err).Error())
			}
		} else {
			err := serverConnection.send(message.NewResponse(message.GetTopic(), client.name, response))
			if err != nil {
				client.logger.Log(Error.New("Failed to send received response to message broker server", err).Error())
			}
		}
	} else {
		err := client.handleAsyncMessage(message)
		if !client.handleServerMessagesConcurrently {
			client.handleServerMessagesConcurrentlyMutex.Unlock()
		}
		if err != nil {
			client.logger.Log(Error.New("Error handling message", err).Error())
		}
	}
}

func (client *Client) handleSyncMessage(message *Message.Message) (string, error) {
	client.mapOperationMutex.Lock()
	messageHandler := client.application.GetSyncMessageHandlers()[message.GetTopic()]
	client.mapOperationMutex.Unlock()
	if messageHandler == nil {
		return "", Error.New("No message handler for topic \""+message.GetTopic()+"\" on \""+client.name+"\"", nil)
	}
	response, err := messageHandler(message)
	if err != nil {
		return "", Error.New("Error handling message", err)
	} else {
		return response, nil
	}
}

func (client *Client) handleAsyncMessage(message *Message.Message) error {
	client.mapOperationMutex.Lock()
	messageHandler := client.application.GetAsyncMessageHandlers()[message.GetTopic()]
	client.mapOperationMutex.Unlock()
	if messageHandler == nil {
		return Error.New("No message handler for topic \""+message.GetTopic()+"\" on \""+client.name+"\"", nil)
	}
	err := messageHandler(message)
	if err != nil {
		return Error.New("Error handling message", err)
	}
	return nil
}
