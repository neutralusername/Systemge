package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

// sync responses are received in this function
func (node *Node) handleOutgoingConnectionMessages(outgoingConnection *outgoingConnection) {
	if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
		infoLogger.Log(Error.New("Starting handling of messages from outgoing node connection \""+outgoingConnection.name+"\"", nil).Error())
	}
	systemge_ := node.systemge
	for {
		systemge := node.systemge
		if systemge == nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Aborting handling of messages from outgoing node connection \""+outgoingConnection.name+"\" because systemge is nil likely due to node being stopped", nil).Error())
			}
			return
		}
		if systemge != systemge_ {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Aborting handling of messages from outgoing node connection \""+outgoingConnection.name+"\" because systemge has changed likely due to node restart", nil).Error())
			}
			return
		}
		messageBytes, err := systemge.receiveOutgoingConnection(outgoingConnection)
		if err != nil {
			if errorLogger := node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Failed to receive message from outgoing node connection \""+outgoingConnection.name+"\" likely due to connection loss", err).Error())
			}
			systemge.outgoingConnectionMutex.Lock()
			delete(systemge.outgoingConnections, outgoingConnection.name)
			for _, topic := range outgoingConnection.topics {
				topicResolutions := systemge.topicResolutions[topic]
				if topicResolutions != nil {
					delete(topicResolutions, outgoingConnection.name)
					if len(topicResolutions) == 0 {
						delete(systemge.topicResolutions, topic)
					}
				}
			}
			systemge.outgoingConnectionMutex.Unlock()
			go node.outgoingConnectionLoop(outgoingConnection.endpointConfig)
			return
		}
		message, err := Message.Deserialize(messageBytes)
		if err != nil {
			systemge.invalidMessagesFromOutgoingConnections.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\"", err).Error())
			}
			continue
		}
		if len(message.GetSyncTokenToken()) == 0 {
			systemge.invalidMessagesFromOutgoingConnections.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Received sync response from outgoing node connection", nil).Error())
			}
			continue
		}
		if err := systemge.validateMessage(message); err != nil {
			systemge.invalidMessagesFromOutgoingConnections.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to validate message", err).Error())
			}
			continue
		}
		err = systemge.handleSyncResponse(outgoingConnection.name, message)
		if err != nil {
			systemge.invalidMessagesFromOutgoingConnections.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to handle sync response from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
			}
			continue
		}
	}
}

func (systemge *systemgeComponent) handleSyncResponse(outgoingConnectionName string, message *Message.Message) error {
	systemge.syncRequestMutex.Lock()
	defer systemge.syncRequestMutex.Unlock()
	syncResponseChannel := systemge.syncRequestChannels[message.GetSyncTokenToken()]
	if syncResponseChannel == nil {
		return Error.New("Received sync response for unknown token", nil)
	}
	if message.GetTopic() == Message.TOPIC_SUCCESS {
		systemge.incomingSyncSuccessResponses.Add(1)
	} else {
		systemge.incomingSyncFailureResponses.Add(1)
	}
	systemge.incomingSyncResponses.Add(1)
	delete(systemge.syncRequestChannels, message.GetSyncTokenToken())
	syncResponseChannel.channel <- &SyncResponse{
		origin:          outgoingConnectionName,
		responseMessage: message,
	}
	return nil
}
