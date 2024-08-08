package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

// handles incoming messages from a incoming connection one at a time until the receive operation fails
// either due to connection loss or closure of the listener due to systemge stop
func (systemge *systemgeComponent) handleIncomingConnectionMessages(incomingConnection *incomingConnection) {
	if infoLogger := systemge.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting message handler for incoming node connection \""+incomingConnection.name+"\"", nil).Error())
	}
	for {
		messageBytes, err := incomingConnection.receiveMessage(systemge.config.TcpBufferBytes, systemge.config.IncomingMessageByteLimit)
		if err != nil {
			if warningLogger := systemge.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from incoming node connection \""+incomingConnection.name+"\" likely due to connection loss", err).Error())
			}
			incomingConnection.netConn.Close()
			if incomingConnection.rateLimiterBytes != nil {
				incomingConnection.rateLimiterBytes.Stop()
			}
			if incomingConnection.rateLimiterMsgs != nil {
				incomingConnection.rateLimiterMsgs.Stop()
			}
			close(incomingConnection.stopChannel)
			systemge.incomingConnectionMutex.Lock()
			delete(systemge.incomingConnections, incomingConnection.name)
			systemge.incomingConnectionMutex.Unlock()
			return
		}
		if err := systemge.checkRateLimits(incomingConnection, messageBytes); err != nil {
			if warningLogger := systemge.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Rejected message from incoming node connection \""+incomingConnection.name+"\"", err).Error())
			}
			continue
		}
		message, err := Message.Deserialize(messageBytes, incomingConnection.name)
		if err != nil {
			systemge.invalidMessagesFromIncomingConnections.Add(1)
			if warningLogger := systemge.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\" from incoming node connection \""+incomingConnection.name+"\"", err).Error())
			}
			continue
		}
		if err := systemge.validateMessage(message); err != nil {
			systemge.invalidMessagesFromIncomingConnections.Add(1)
			if warningLogger := systemge.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to validate message \""+string(messageBytes)+"\" from incoming node connection \""+incomingConnection.name+"\"", err).Error())
			}
			continue
		}
		if message.GetSyncTokenToken() == "" {
			systemge.incomingAsyncMessageBytesReceived.Add(uint64(len(messageBytes)))
			systemge.incomingAsyncMessages.Add(1)
			err := systemge.handleAsyncMessage(message)
			if err != nil {
				systemge.invalidMessagesFromIncomingConnections.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle async messag with topic \""+message.GetTopic()+"\" from incoming node connection \""+incomingConnection.name+"\"", err).Error())
				}
			} else {
				if infoLogger := systemge.infoLogger; infoLogger != nil {
					infoLogger.Log(Error.New("Handled async message with topic \""+message.GetTopic()+"\" from incoming node connection \""+incomingConnection.name+"\"", nil).Error())
				}
			}
		} else {
			systemge.incomingSyncRequestBytesReceived.Add(uint64(len(messageBytes)))
			systemge.incomingSyncRequests.Add(1)
			responsePayload, err := systemge.handleSyncRequest(message)
			if err != nil {
				systemge.invalidMessagesFromIncomingConnections.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to handle sync request with topic \""+message.GetTopic()+"\" from incoming node connection \""+incomingConnection.name+"\"", err).Error())
				}
				if err := systemge.messageIncomingConnection(incomingConnection, message.NewFailureResponse(responsePayload)); err != nil {
					if warningLogger := systemge.warningLogger; warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send failure response to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
				} else {
					systemge.outgoingSyncFailureResponses.Add(1)
				}
			} else {
				if infoLogger := systemge.infoLogger; infoLogger != nil {
					infoLogger.Log(Error.New("Handled sync request with topic \""+message.GetTopic()+"\" from incoming node connection \""+incomingConnection.name+"\" with sync token \""+message.GetSyncTokenToken()+"\"", nil).Error())
				}
				if err := systemge.messageIncomingConnection(incomingConnection, message.NewSuccessResponse(responsePayload)); err != nil {
					if warningLogger := systemge.warningLogger; warningLogger != nil {
						warningLogger.Log(Error.New("Failed to send success response to incoming node connection \""+incomingConnection.name+"\"", err).Error())
					}
				} else {
					systemge.outgoingSyncSuccessResponses.Add(1)
				}
			}
		}
	}
}

func (systemge *systemgeComponent) checkRateLimits(incomingConnection *incomingConnection, messageBytes []byte) error {
	systemge.bytesReceived.Add(uint64(len(messageBytes)))
	if incomingConnection.rateLimiterBytes != nil && !incomingConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		systemge.incomingConnectionRateLimiterBytesExceeded.Add(1)
		return Error.New("Incoming connection rate limiter bytes exceeded", nil)
	}
	if incomingConnection.rateLimiterMsgs != nil && !incomingConnection.rateLimiterMsgs.Consume(1) {
		systemge.incomingConnectionRateLimiterMsgsExceeded.Add(1)
		return Error.New("Incoming connection rate limiter messages exceeded", nil)
	}
	return nil
}
