package Node

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (node *Node) handleSystemgeMessages(brokerConnection *brokerConnection) {
	for brokerConnection.netConn != nil {
		message, err := brokerConnection.receive()
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", err).Error())
			}
			node.handleBrokerDisconnect(brokerConnection)
			return
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Received message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", nil).Error())
		}
		if message.GetSyncResponseToken() != "" {
			err := node.handleSyncResponse(message)
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", err).Error())
				}
			} else {
				if infoLogger := node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", nil).Error())
				}
			}
			continue
		}
		if message.GetSyncRequestToken() != "" {
			response, err := node.handleSyncMessage(message)
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", err).Error())
				}
				err := node.send(brokerConnection, message.NewResponse("error", node.GetName(), Error.New("failed handling message", err).Error()))
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send error response for failed sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", err).Error())
					}
				}
			} else {
				if infoLogger := node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", nil).Error())
				}
				err = node.send(brokerConnection, message.NewResponse(message.GetTopic(), node.GetName(), response))
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send response for sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", err).Error())
					}
				}
			}
			continue
		}
		if node.GetSystemgeComponent().GetSystemgeComponentConfig().HandleMessagesSequentially {
			node.systemgeHandleSequentiallyMutex.Lock()
		}
		err = node.handleAsyncMessage(message)
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to handle message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", err).Error())
			}
		} else {
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Handled message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\" on node \""+node.GetName()+"\"", nil).Error())
			}
			if node.GetSystemgeComponent().GetSystemgeComponentConfig().HandleMessagesSequentially {
				node.systemgeHandleSequentiallyMutex.Unlock()
			}
		}
		if node.GetSystemgeComponent().GetSystemgeComponentConfig().HandleMessagesSequentially {
			node.systemgeHandleSequentiallyMutex.Unlock()
		}
	}
}

func (node *Node) handleAsyncMessage(message *Message.Message) error {
	asyncHandler := node.GetSystemgeComponent().GetAsyncMessageHandlers()[message.GetTopic()]
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
	syncHandler := node.GetSystemgeComponent().GetSyncMessageHandlers()[message.GetTopic()]
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
	node.systemgeMutex.Lock()
	responseChannel := node.systemgeMessagesWaitingForResponse[message.GetSyncResponseToken()]
	node.systemgeMutex.Unlock()
	if responseChannel == nil {
		return Error.New("Unknown sync response token", nil)
	}
	responseChannel <- message
	return nil
}
