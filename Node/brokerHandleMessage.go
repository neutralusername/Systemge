package Node

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (node *Node) handleBrokerMessages(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		message, err := brokerConnection.receive()
		if err != nil {
			node.config.Logger.Warning(Error.New("Failed to receive message from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
			node.handleBrokerDisconnect(brokerConnection)
			return
		} else {
			node.config.Logger.Info(Error.New("Received message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", nil).Error())
		}
		if message.GetSyncResponseToken() != "" {
			err := node.handleSyncResponse(message)
			if err != nil {
				node.config.Logger.Warning(Error.New("Failed to handle sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
			} else {
				node.config.Logger.Info(Error.New("Received sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", nil).Error())
			}
			continue
		}
		if message.GetSyncRequestToken() != "" {
			response, err := node.handleSyncMessage(message)
			if err != nil {
				node.config.Logger.Warning(Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
				err := brokerConnection.send(message.NewResponse("error", node.config.Name, Error.New("failed handling message", err).Error()))
				if err != nil {
					node.config.Logger.Warning(Error.New("Failed to send error response for failed sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
				}
			} else {
				node.config.Logger.Info(Error.New("Handled sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", nil).Error())
				err = brokerConnection.send(message.NewResponse(message.GetTopic(), node.config.Name, response))
				if err != nil {
					node.config.Logger.Warning(Error.New("Failed to send response for sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
				}
			}
			continue
		}
		if node.application.GetApplicationConfig().HandleMessagesSequentially {
			node.handleMessagesSequentiallyMutex.Lock()
		}
		err = node.handleAsyncMessage(message)
		if err != nil {
			node.config.Logger.Warning(Error.New("Failed to handle message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
		} else {
			node.config.Logger.Info(Error.New("Handled message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", nil).Error())
		}
		if node.application.GetApplicationConfig().HandleMessagesSequentially {
			node.handleMessagesSequentiallyMutex.Unlock()
		}
	}
}

func (node *Node) handleAsyncMessage(message *Message.Message) error {
	asyncHandler := node.application.GetAsyncMessageHandlers()[message.GetTopic()]
	if asyncHandler == nil {
		return Error.New("No handler", nil)
	}
	err := asyncHandler(node, message)
	if err != nil {
		return Error.New("Failed to handle message", err)
	}
	return nil
}

func (node *Node) handleSyncMessage(message *Message.Message) (string, error) {
	syncHandler := node.application.GetSyncMessageHandlers()[message.GetTopic()]
	if syncHandler == nil {
		return "", Error.New("No handler", nil)
	}
	response, err := syncHandler(node, message)
	if err != nil {
		return "", Error.New("Failed to handle message", err)
	}
	return response, nil
}

func (node *Node) handleSyncResponse(message *Message.Message) error {
	node.mutex.Lock()
	responseChannel := node.messagesWaitingForResponse[message.GetSyncResponseToken()]
	node.mutex.Unlock()
	if responseChannel == nil {
		return Error.New("Unknown sync response token", nil)
	}
	responseChannel <- message
	return nil
}
