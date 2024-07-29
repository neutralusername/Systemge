package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (node *Node) handleBrokerConnectionMessages(brokerConnection *brokerConnection) {
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Handling messages from broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
	}
	for {
		systemge := node.systemge
		if systemge == nil {
			return
		}
		brokerConnection.receiveMutex.Lock()
		messageBytes, bytesReceived, err := Tcp.Receive(brokerConnection.netConn, 0, 0)
		brokerConnection.receiveMutex.Unlock()
		if err != nil {
			close(brokerConnection.closeChannel)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
			}
			for _, topic := range systemge.cleanUpDisconnectedBrokerConnection(brokerConnection) {
				go func(topic string) {
					err := node.subscribeLoop(topic, systemge.application.GetSystemgeComponentConfig().MaxSubscribeAttempts)
					if err != nil {
						if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
							warningLogger.Log(Error.New("Failed to subscribe for topic \""+topic+"\"", err).Error())
						}
						if node.systemge == systemge {
							err := node.stop(true)
							if err != nil {
								if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
									warningLogger.Log(Error.New("Failed to stop node due to failed subscription for topic \""+topic+"\"", err).Error())
								}
							}
						}
					}
				}(topic)
			}
			return
		}
		systemge.bytesReceivedCounter.Add(bytesReceived)
		message := Message.Deserialize(messageBytes)
		if message == nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to deserialize message from broker \""+brokerConnection.endpoint.Address+"\" with "+string(messageBytes)+" bytes", nil).Error())
			}
			continue
		}
		if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Received message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
		}
		if message.GetSyncResponseToken() != "" {
			err := node.handleSyncResponse(message)
			if err != nil {
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
				}
			} else {
				systemge.incomingSyncResponseCounter.Add(1)
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled sync response with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncResponseToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
				}
			}
			continue
		}
		if message.GetSyncRequestToken() != "" {
			systemge.incomingSyncRequestCounter.Add(1)
			response, err := node.handleSyncMessage(message)
			if err != nil {
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
				}
				bytesSent, err := brokerConnection.send(systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, message.NewResponse("error", node.GetName(), Error.New("failed handling message", err).Error()).Serialize())
				if err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send error response for failed sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
					}
				} else {
					systemge.bytesSentCounter.Add(bytesSent)
					systemge.outgoingSyncResponseCounter.Add(1)
				}
			} else {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
				}
				bytesSent, err := brokerConnection.send(systemge.application.GetSystemgeComponentConfig().TcpTimeoutMs, message.NewResponse(message.GetTopic(), node.GetName(), response).Serialize())
				if err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send response for sync request with topic \""+message.GetTopic()+"\" and token \""+message.GetSyncRequestToken()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
					}
				} else {
					systemge.bytesSentCounter.Add(bytesSent)
					systemge.outgoingSyncResponseCounter.Add(1)
				}
			}
			continue
		}
		systemge.incomingAsyncMessageCounter.Add(1)
		if systemge.application.GetSystemgeComponentConfig().HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Lock()
		}
		err = node.handleAsyncMessage(message)
		if systemge.application.GetSystemgeComponentConfig().HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Unlock()
		}
		if err != nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to handle message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", err).Error())
			}
		} else {
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log(Error.New("Handled message with topic \""+message.GetTopic()+"\" from broker \""+brokerConnection.endpoint.Address+"\"", nil).Error())
			}
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
	node.systemge.messagesWaitingForResponseMutex.Lock()
	responseChannel := node.systemge.messagesWaitingForResponse[message.GetSyncResponseToken()]
	node.systemge.messagesWaitingForResponseMutex.Unlock()
	if responseChannel == nil {
		return Error.New("Unknown sync response token", nil)
	}
	responseChannel <- message
	return nil
}
