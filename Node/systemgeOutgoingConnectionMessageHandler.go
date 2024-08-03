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
			if outgoingConnection.rateLimiterBytes != nil {
				outgoingConnection.rateLimiterBytes.Stop()
			}
			if outgoingConnection.rateLimiterMsgs != nil {
				outgoingConnection.rateLimiterMsgs.Stop()
			}
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
		go func() {
			if outgoingConnection.rateLimiterBytes != nil && !outgoingConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
				systemge.outgoingConnectionRateLimiterBytesExceeded.Add(1)
				return
			}
			if outgoingConnection.rateLimiterMsgs != nil && !outgoingConnection.rateLimiterMsgs.Consume(1) {
				systemge.outgoingConnectionRateLimiterMsgsExceeded.Add(1)
				return
			}
			message, err := Message.Deserialize(messageBytes, outgoingConnection.name)
			if err != nil {
				systemge.invalidMessagesFromOutgoingConnections.Add(1)
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
				return
			}
			if len(message.GetSyncTokenToken()) == 0 {
				systemge.invalidMessagesFromOutgoingConnections.Add(1)
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Received async message from outgoing node connection \""+outgoingConnection.name+"\" (which goes against protocol)", nil).Error())
				}
				return
			}
			systemge.incomingSyncResponseBytesReceived.Add(uint64(len(messageBytes)))
			if err := systemge.validateMessage(message); err != nil {
				systemge.invalidMessagesFromOutgoingConnections.Add(1)
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to validate message \""+string(messageBytes)+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
				return
			}
			err = systemge.handleSyncResponse(message)
			if err != nil {
				systemge.invalidMessagesFromOutgoingConnections.Add(1)
				if warningLogger := node.GetInternalWarningError(); warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle sync response from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
				return
			} else {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("Handled sync response from outgoing node connection \"" + outgoingConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
				}
			}
		}()
	}
}
