package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

func (node *Node) handleIncomingConnectionMessages(incomingConnection *incomingConnection) {
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Starting handling of messages from incoming node connection \""+incomingConnection.name+"\"", nil).Error())
	}
	systemge_ := node.systemge
	for {
		systemge := node.systemge
		if systemge == nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Aborting handling of messages from incoming node connection \""+incomingConnection.name+"\" because systemge is nil likely due to node being stopped", nil).Error())
			}
			return
		}
		if systemge != systemge_ {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Aborting handling of messages from incoming node connection \""+incomingConnection.name+"\" because systemge has changed likely due to node restart", nil).Error())
			}
			return
		}
		messageBytes, err := systemge.receiveFromIncomingConnection(incomingConnection)
		if err != nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from incoming node connection \""+incomingConnection.name+"\" likely due to connection loss", err).Error())
			}
			incomingConnection.netConn.Close()
			systemge.removeIncomingConnection(incomingConnection)
			return
		}
		message, err := Message.Deserialize(messageBytes)
		if err != nil {
			systemge.invalidMessagesFromIncomingConnections.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\" from incoming node connection \""+incomingConnection.name+"\"", err).Error())
			}
			continue
		}
		if err := systemge.validateMessage(message); err != nil {
			systemge.invalidMessagesFromIncomingConnections.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to validate message \""+string(messageBytes)+"\" from incoming node connection \""+incomingConnection.name+"\"", err).Error())
			}
			continue
		}
		if systemge.config.HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Lock()
		}
		if message.GetSyncTokenToken() == "" {
			systemge.incomingAsyncMessageBytesReceived.Add(uint64(len(messageBytes)))
			systemge.incomingAsyncMessages.Add(1)
			err := systemge.handleAsyncMessage(node, message)
			if err != nil {
				systemge.invalidMessagesFromIncomingConnections.Add(1)
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle async messag with topic \""+message.GetTopic()+"\" from incoming node connection \""+incomingConnection.name+"\"", err).Error())
				}
			} else {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled async message with topic \""+message.GetTopic()+"\" from incoming node connection \""+incomingConnection.name+"\"", nil).Error())
				}
			}
		} else {
			systemge.incomingSyncRequestBytesReceived.Add(uint64(len(messageBytes)))
			systemge.incomingSyncRequests.Add(1)
			responsePayload, err := systemge.handleSyncRequest(node, message)
			if err != nil {
				systemge.invalidMessagesFromIncomingConnections.Add(1)
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" from incoming node connection \""+incomingConnection.name+"\"", err).Error())
				}
				if err := systemge.messageIncomingConnection(incomingConnection, message.NewFailureResponse(responsePayload)); err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send failure response to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
				} else {
					systemge.outgoingSyncFailureResponses.Add(1)
				}
			} else {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log(Error.New("Handled sync request with topic \""+message.GetTopic()+"\" from incoming node connection \""+incomingConnection.name+"\" with sync token \""+message.GetSyncTokenToken()+"\"", nil).Error())
				}
				if err := systemge.messageIncomingConnection(incomingConnection, message.NewSuccessResponse(responsePayload)); err != nil {
					if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send success response to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
				} else {
					systemge.outgoingSyncSuccessResponses.Add(1)
				}
			}
		}
		if systemge.config.HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Unlock()
		}
	}
}

func (systemge *systemgeComponent) handleSyncRequest(node *Node, message *Message.Message) (string, error) {
	systemge.syncMessageHandlerMutex.Lock()
	syncMessageHandler := systemge.application.GetSyncMessageHandlers()[message.GetTopic()]
	systemge.syncMessageHandlerMutex.Unlock()
	if syncMessageHandler == nil {
		return "Not responsible for topic \"" + message.GetTopic() + "\"", Error.New("Received sync request with topic \""+message.GetTopic()+"\" for which no handler is registered", nil)
	}
	responsePayload, err := syncMessageHandler(node, message)
	if err != nil {
		return err.Error(), Error.New("Sync message handler for topic \""+message.GetTopic()+"\" failed returned error", err)
	}
	return responsePayload, nil
}

func (systemge *systemgeComponent) handleAsyncMessage(node *Node, message *Message.Message) error {
	systemge.asyncMessageHandlerMutex.Lock()
	asyncMessageHandler := systemge.application.GetAsyncMessageHandlers()[message.GetTopic()]
	systemge.asyncMessageHandlerMutex.Unlock()
	if asyncMessageHandler == nil {
		return Error.New("Received async message with topic \""+message.GetTopic()+"\" for which no handler is registered", nil)
	}
	err := asyncMessageHandler(node, message)
	if err != nil {
		return Error.New("Async message handler for topic \""+message.GetTopic()+"\" failed returned error", err)
	}
	return nil
}
