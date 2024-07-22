package Node

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (node *Node) handleBrokerConnectionMessages(brokerConnection *brokerConnection) {
	for {
		systemge := node.systemge
		if systemge == nil {
			return
		}
		message, err := brokerConnection.receive()
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
			}
			brokerConnection.close()
			removedSubscribedTopics := systemge.cleanUpDisconnectedBrokerConnection(brokerConnection)
			for _, topic := range removedSubscribedTopics {
				go func() {
					err := node.subscribeLoop(topic, systemge.application.GetSystemgeComponentConfig().MaxSubscribeAttempts)
					if err != nil {
						if warningLogger := node.GetWarningLogger(); warningLogger != nil {
							warningLogger.Log(Error.New("Failed to subscribe for topic \""+topic+"\"", err).Error())
						}
						err := node.stop(true)
						if err != nil {
							if warningLogger := node.GetWarningLogger(); warningLogger != nil {
								warningLogger.Log(Error.New("Failed to stop node due to failed subscription for topic \""+topic+"\"", err).Error())
							}
						}
					}
				}()
			}
			return
		}
		if infoLogger := node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Received message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
		}
		if message.GetSyncResponseToken() != "" {
			err := node.handleSyncResponse(message)
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
				}
			} else {
				if infoLogger := node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
				}
			}
			continue
		}
		if message.GetSyncRequestToken() != "" {
			response, err := node.handleSyncMessage(message)
			if err != nil {
				if warningLogger := node.GetWarningLogger(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
				}
				err := brokerConnection.send(systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, message.NewResponse("error", node.GetName(), Error.New("failed handling message", err).Error()))
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send error response for failed sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
					}
				}
			} else {
				if infoLogger := node.GetInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
				}
				err = brokerConnection.send(systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, message.NewResponse(message.GetTopic(), node.GetName(), response))
				if err != nil {
					if warningLogger := node.GetWarningLogger(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send response for sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
					}
				}
			}
			continue
		}
		if systemge.application.GetSystemgeComponentConfig().HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Lock()
		}
		err = node.handleAsyncMessage(message)
		if err != nil {
			if warningLogger := node.GetWarningLogger(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to handle message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
			}
		} else {
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Handled message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
			}
			if systemge.application.GetSystemgeComponentConfig().HandleMessagesSequentially {
				systemge.handleSequentiallyMutex.Unlock()
			}
		}
		if systemge.application.GetSystemgeComponentConfig().HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Unlock()
		}
	}
}

func (node *Node) handleAsyncMessage(message *Message.Message) error {
	node.systemge.asyncMessageHandlerMutex.Lock()
	asyncHandler := node.systemge.application.GetAsyncMessageHandlers()[message.GetTopic()]
	node.systemge.asyncMessageHandlerMutex.Unlock()
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
	node.systemge.syncMessageHandlerMutex.Lock()
	syncHandler := node.systemge.application.GetSyncMessageHandlers()[message.GetTopic()]
	node.systemge.syncMessageHandlerMutex.Unlock()
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
	node.systemge.mutex.Lock()
	responseChannel := node.systemge.messagesWaitingForResponse[message.GetSyncResponseToken()]
	node.systemge.mutex.Unlock()
	if responseChannel == nil {
		return Error.New("Unknown sync response token", nil)
	}
	responseChannel <- message
	return nil
}
