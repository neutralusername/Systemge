package Node

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

func (client *SystemgeClient) handleOutgoingConnectionMessages(outgoingConnection *serverConnection) {
	if infoLogger := client.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting handling of messages from outgoing node connection \""+outgoingConnection.name+"\"", nil).Error())
	}
	for {
		messageBytes, err := outgoingConnection.receiveMessage(client.config.TcpBufferBytes, client.config.IncomingMessageByteLimit)
		if err != nil {
			if warningLogger := client.warningLogger; warningLogger != nil {
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
			client.serverConnectionMutex.Lock()
			delete(client.serverConnections, outgoingConnection.endpointConfig.Address)
			outgoingConnection.topicsMutex.Lock()
			for topic := range outgoingConnection.topics {
				topicResolutions := client.topicResolutions[topic]
				if topicResolutions != nil {
					delete(topicResolutions, outgoingConnection.name)
					if len(topicResolutions) == 0 {
						delete(client.topicResolutions, topic)
					}
				}
			}
			outgoingConnection.topicsMutex.Unlock()
			if !outgoingConnection.isTransient {
				go func() {
					err := client.attemptOutgoingConnection(outgoingConnection.endpointConfig, false)
					if err != nil {
						if errorLogger := client.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed to reconnect to endpoint \""+outgoingConnection.endpointConfig.Address+"\"", err).Error())
						}
						if client.config.StopAfterOutgoingConnectionLoss {
							client.Stop()
						}
					}
				}()
			}
			client.serverConnectionMutex.Unlock()
			return
		}
		go func(messageBytes []byte) {
			client.bytesReceived.Add(uint64(len(messageBytes)))
			if outgoingConnection.rateLimiterBytes != nil && !outgoingConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
				client.byteRateLimiterExceeded.Add(1)
				return
			}
			if outgoingConnection.rateLimiterMsgs != nil && !outgoingConnection.rateLimiterMsgs.Consume(1) {
				client.messageRateLimiterExceeded.Add(1)
				return
			}
			message, err := Message.Deserialize(messageBytes, outgoingConnection.name)
			if err != nil {
				client.invalidMessagesReceived.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
				return
			}
			if len(message.GetSyncTokenToken()) == 0 {
				if message.GetTopic() == Message.TOPIC_ADDTOPIC {
					outgoingConnection.topicsMutex.Lock()
					outgoingConnection.topics[message.GetPayload()] = true
					outgoingConnection.topicsMutex.Unlock()
					client.topicAddReceived.Add(1)
				} else if message.GetTopic() == Message.TOPIC_REMOVETOPIC {
					outgoingConnection.topicsMutex.Lock()
					delete(outgoingConnection.topics, message.GetPayload())
					outgoingConnection.topicsMutex.Unlock()
					client.topicRemoveReceived.Add(1)
				} else {
					client.invalidMessagesReceived.Add(1)
					if warningLogger := client.warningLogger; warningLogger != nil {
						warningLogger.Log(Error.New("Received async message from outgoing node connection \""+outgoingConnection.name+"\" (which goes against protocol)", nil).Error())
					}
				}
				return
			}
			client.syncResponseBytesReceived.Add(uint64(len(messageBytes)))
			if err := client.validateMessage(message); err != nil {
				client.invalidMessagesReceived.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to validate message \""+string(messageBytes)+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
				return
			}
			syncResponseChannel := client.getResponseChannel(message.GetSyncTokenToken())
			if syncResponseChannel == nil {
				client.invalidMessagesReceived.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to get sync response channel for sync token \""+message.GetSyncTokenToken()+"\" from outgoing node connection \""+outgoingConnection.name+"\"", nil).Error())
				}
				return
			}
			if err = syncResponseChannel.addResponse(message); err != nil {
				client.invalidMessagesReceived.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to add sync response to sync response channel for sync token \""+message.GetSyncTokenToken()+"\" from outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
				}
			} else {
				if message.GetTopic() == Message.TOPIC_SUCCESS {
					client.syncSuccessResponsesReceived.Add(1)
				} else {
					client.syncFailureResponsesReceived.Add(1)
				}
			}
		}(messageBytes)
	}
}

func (client *SystemgeClient) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := client.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncTokenToken()) > maxSyncTokenSize {
		return Error.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Error.New("Message missing topic", nil)
	}
	if maxTopicSize := client.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Error.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := client.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Error.New("Message payload exceeds maximum size", nil)
	}
	return nil
}
