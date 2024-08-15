package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
)

func (client *SystemgeClient) handleServerConnectionMessages(serverConnection *serverConnection) {
	if infoLogger := client.infoLogger; infoLogger != nil {
		infoLogger.Log(Error.New("Starting handling of messages from server connection \""+serverConnection.name+"\"", nil).Error())
	}
	for {
		messageBytes, err := serverConnection.receiveMessage(client.config.TcpBufferBytes, client.config.IncomingMessageByteLimit)
		if err != nil {
			if warningLogger := client.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Failed to receive message from server connection \""+serverConnection.name+"\"", err).Error())
			}
			if client.GetStatus() == Status.STARTED {
				if err := client.Stop(); err != nil {
					if errorLogger := client.errorLogger; errorLogger != nil {
						errorLogger.Log(Error.New("Failed to stop SystemgeClient", err).Error())
					}
				}
				if client.config.RestartAfterConnectionLoss {
					if err := client.Start(); err != nil {
						if errorLogger := client.errorLogger; errorLogger != nil {
							errorLogger.Log(Error.New("Failed to restart SystemgeClient after connection loss", err).Error())
						}
					}
				}
			}
			return
		}
		go func(messageBytes []byte) {
			client.bytesReceived.Add(uint64(len(messageBytes)))
			if serverConnection.rateLimiterBytes != nil && !serverConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
				client.byteRateLimiterExceeded.Add(1)
				return
			}
			if serverConnection.rateLimiterMsgs != nil && !serverConnection.rateLimiterMsgs.Consume(1) {
				client.messageRateLimiterExceeded.Add(1)
				return
			}
			message, err := Message.Deserialize(messageBytes, serverConnection.name)
			if err != nil {
				client.invalidMessagesReceived.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to deserialize message \""+string(messageBytes)+"\" from server connection \""+serverConnection.name+"\"", err).Error())
				}
				return
			}
			if len(message.GetSyncTokenToken()) == 0 {
				client.invalidMessagesReceived.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Received async message from server connection \""+serverConnection.name+"\" (which goes against protocol)", nil).Error())
				}
				return
			}
			if err := client.validateMessage(message); err != nil {
				client.invalidMessagesReceived.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to validate message \""+string(messageBytes)+"\" from server connection \""+serverConnection.name+"\"", err).Error())
				}
				return
			}
			if err := client.addSyncResponse(message); err != nil {
				client.invalidMessagesReceived.Add(1)
				if warningLogger := client.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to add sync response message \""+string(messageBytes)+"\" from server connection \""+serverConnection.name+"\"", err).Error())
				}
			} else {
				client.syncResponseBytesReceived.Add(uint64(len(messageBytes)))
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
