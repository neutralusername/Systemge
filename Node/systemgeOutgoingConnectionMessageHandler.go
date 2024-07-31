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
		messageBytes, err := systemge.receiveFromOutgoingConnection(outgoingConnection)
		if err != nil {
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
			}
			outgoingConnection.netConn.Close()

			defer systemge.outgoingConnectionMutex.Unlock()
			systemge.outgoingConnectionMutex.Lock()
			defer delete(systemge.outgoingConnections, outgoingConnection.endpointConfig.Address)
			for _, topic := range outgoingConnection.topics {
				topicResolutions := systemge.topicResolutions[topic]
				if topicResolutions != nil {
					delete(topicResolutions, outgoingConnection.endpointConfig.Address)
					if len(topicResolutions) == 0 {
						delete(systemge.topicResolutions, topic)
					}
				}
			}
			if !outgoingConnection.transient {
				go node.StartOutgoingConnectionLoop(outgoingConnection.endpointConfig)
			}
			return
		}
		message, err := Message.Deserialize(messageBytes)
		if err != nil {
			systemge.invalidMessagesFromOutgoingConnections.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
			}
			continue
		}
		if len(message.GetSyncTokenToken()) == 0 {
			systemge.invalidMessagesFromOutgoingConnections.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Received async message from outgoing node connection \""+outgoingConnection.name+"\" (which goes against protocol)", nil).Error())
			}
			continue
		}
		if err := systemge.validateMessage(message); err != nil {
			systemge.invalidMessagesFromOutgoingConnections.Add(1)
			if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
				warningLogger.Log(Error.New("Failed to validate message \""+string(messageBytes)+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
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
		} else {
			if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
				infoLogger.Log("Handled sync response from outgoing node connection \"" + outgoingConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
			}
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
	if syncResponseChannel.responseCount >= systemge.config.SyncResponseLimit {
		return Error.New("Sync response limit reached", nil)
	}
	if message.GetTopic() == Message.TOPIC_SUCCESS {
		systemge.incomingSyncSuccessResponses.Add(1)
	} else {
		systemge.incomingSyncFailureResponses.Add(1)
	}
	syncResponseChannel.responseCount++
	systemge.incomingSyncResponses.Add(1)
	delete(systemge.syncRequestChannels, message.GetSyncTokenToken())
	go syncResponseChannel.addResponse(&SyncResponse{
		origin:          outgoingConnectionName,
		responseMessage: message,
	})
	return nil
}
