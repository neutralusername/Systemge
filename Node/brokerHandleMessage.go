package Node

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (node *Node) handleBrokerMessages(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		message, err := brokerConnection.receive()
		if err != nil {
			if node.IsStarted() {
				node.logger.Log(Error.New("failed to receive message from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
			}
			node.handleBrokerDisconnect(brokerConnection)
			return
		}
		if message.GetSyncResponseToken() != "" {
			err := node.handleSyncResponse(message)
			if err != nil {
				node.logger.Log(Error.New("failed to handle sync response with token \""+message.GetSyncResponseToken()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
			}
		} else {
			err := node.handleMessage(message, brokerConnection)
			if err != nil {
				node.logger.Log(Error.New("failed to handle message from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
			}
		}
	}
}

func (node *Node) handleSyncResponse(message *Message.Message) error {
	node.mutex.Lock()
	responseChannel := node.messagesWaitingForResponse[message.GetSyncResponseToken()]
	node.mutex.Unlock()
	if responseChannel == nil {
		return Error.New("unknown sync response token", nil)
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
			errResponse := brokerConnection.send(message.NewResponse("error", node.config.Name, Error.New("failed handling message", err).Error()))
			if errResponse != nil {
				node.logger.Log(Error.New("failed to send error response for failed sync request with token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", errResponse).Error())
			}
			return Error.New("failed handling sync request with token \""+message.GetSyncRequestToken()+"\"", err)
		}
		err = brokerConnection.send(message.NewResponse(message.GetTopic(), node.config.Name, response))
		if err != nil {
			node.logger.Log(Error.New("failed to send response for sync request with token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.GetAddress()+"\" on node \""+node.config.Name+"\"", err).Error())
		}
	} else {
		err := node.handleAsyncMessage(message)
		if err != nil {
			return Error.New("failed handling message", err)
		}
	}
	return nil
}

func (node *Node) handleSyncMessage(message *Message.Message) (string, error) {
	syncHandler := node.application.GetSyncMessageHandlers()[message.GetTopic()]
	if syncHandler == nil {
		return "", Error.New("no handler for topic \""+message.GetTopic()+"\"", nil)
	}
	response, err := syncHandler(node, message)
	if err != nil {
		return "", Error.New("failed handling message with topic \""+message.GetTopic()+"\"", err)
	}
	return response, nil
}

func (node *Node) handleAsyncMessage(message *Message.Message) error {
	asyncHandler := node.application.GetAsyncMessageHandlers()[message.GetTopic()]
	if asyncHandler == nil {
		return Error.New("no handler for topic \""+message.GetTopic()+"\"", nil)
	}
	err := asyncHandler(node, message)
	if err != nil {
		return Error.New("failed handling message with topic \""+message.GetTopic()+"\"", err)
	}
	return nil
}
