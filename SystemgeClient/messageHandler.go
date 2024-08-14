package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

func (systemge *SystemgeClient) handleOutgoingConnectionMessages(outgoingConnection *outgoingConnection) {
	if infoLogger := systemge.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting handling of messages from outgoing node connection \""+outgoingConnection.name+"\"", nil).Error())
	}
	for {
		messageBytes, err := outgoingConnection.receiveMessage(systemge.config.TcpBufferBytes, systemge.config.IncomingMessageByteLimit)
		if err != nil {
			if warningLogger := systemge.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
			}
			outgoingConnection.netConn.Close()
			if outgoingConnection.rateLimiterBytes != nil {
				outgoingConnection.rateLimiterBytes.Stop()
			}
			if outgoingConnection.rateLimiterMsgs != nil {
				outgoingConnection.rateLimiterMsgs.Stop()
			}
			close(outgoingConnection.stopChannel)
			systemge.outgoingConnectionMutex.Lock()
			delete(systemge.outgoingConnections, outgoingConnection.endpointConfig.Address)
			outgoingConnection.topicsMutex.Lock()
			for topic := range outgoingConnection.topics {
				topicResolutions := systemge.topicResolutions[topic]
				if topicResolutions != nil {
					delete(topicResolutions, outgoingConnection.name)
					if len(topicResolutions) == 0 {
						delete(systemge.topicResolutions, topic)
					}
				}
			}
			outgoingConnection.topicsMutex.Unlock()
			if !outgoingConnection.isTransient {
				go func() {
					err := systemge.attemptOutgoingConnection(outgoingConnection.endpointConfig, false)
					if err != nil {
						if errorLogger := systemge.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed to reconnect to endpoint \""+outgoingConnection.endpointConfig.Address+"\"", err).Error())
						}
						if systemge.config.StopAfterOutgoingConnectionLoss {
							systemge.stopNode()
						}
					}
				}()
			}
			systemge.outgoingConnectionMutex.Unlock()
			return
		}
		go func(messageBytes []byte) {
			systemge.bytesReceived.Add(uint64(len(messageBytes)))
			if outgoingConnection.rateLimiterBytes != nil && !outgoingConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
				systemge.byteRateLimiterExceeded.Add(1)
				return
			}
			if outgoingConnection.rateLimiterMsgs != nil && !outgoingConnection.rateLimiterMsgs.Consume(1) {
				systemge.messageRateLimiterExceeded.Add(1)
				return
			}
			message, err := Message.Deserialize(messageBytes, outgoingConnection.name)
			if err != nil {
				systemge.invalidMessagesReceived.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
				return
			}
			if len(message.GetSyncTokenToken()) == 0 {
				if message.GetTopic() == TOPIC_ADDTOPIC {
					outgoingConnection.topicsMutex.Lock()
					outgoingConnection.topics[message.GetPayload()] = true
					outgoingConnection.topicsMutex.Unlock()
					systemge.topicAddReceived.Add(1)
				} else if message.GetTopic() == TOPIC_REMOVETOPIC {
					outgoingConnection.topicsMutex.Lock()
					delete(outgoingConnection.topics, message.GetPayload())
					outgoingConnection.topicsMutex.Unlock()
					systemge.topicRemoveReceived.Add(1)
				} else {
					systemge.invalidMessagesReceived.Add(1)
					if warningLogger := systemge.warningLogger; warningLogger != nil {
						warningLogger.Log(Error.New("Received async message from outgoing node connection \""+outgoingConnection.name+"\" (which goes against protocol)", nil).Error())
					}
				}
				return
			}
			systemge.syncResponseBytesReceived.Add(uint64(len(messageBytes)))
			if err := systemge.validateMessage(message); err != nil {
				systemge.invalidMessagesReceived.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to validate message \""+string(messageBytes)+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
				return
			}
			syncResponseChannel := systemge.getResponseChannel(message.GetSyncTokenToken())
			if syncResponseChannel == nil {
				systemge.invalidMessagesReceived.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to get sync response channel for sync token \""+message.GetSyncTokenToken()+"\" from outgoing node connection \""+outgoingConnection.name+"\"", nil).Error())
				}
				return
			}
			if err = syncResponseChannel.addResponse(message); err != nil {
				systemge.invalidMessagesReceived.Add(1)
				if warningLogger := systemge.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to add sync response to sync response channel for sync token \""+message.GetSyncTokenToken()+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
			} else {
				if message.GetTopic() == Message.TOPIC_SUCCESS {
					systemge.syncSuccessResponsesReceived.Add(1)
				} else {
					systemge.syncFailureResponsesReceived.Add(1)
				}
			}
		}(messageBytes)
	}
}
