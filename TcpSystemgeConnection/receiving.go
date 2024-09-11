package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpConnection) receiveLoop() {
	if connection.infoLogger != nil {
		connection.infoLogger.Log("Started receiving messages")
	}

	for {
		select {
		case <-connection.closeChannel:
			if connection.infoLogger != nil {
				connection.infoLogger.Log("Stopped receiving messages")
			}
			connection.Close()
			return
		default:
			// block here until we can be sure that we can add a message to the processing channel
			connection.unprocessedMessages.Add(1)
			messageBytes, err := connection.receive()
			if err != nil {
				if connection.warningLogger != nil {
					connection.warningLogger.Log(Error.New("failed to receive message", err).Error())
				}
				connection.unprocessedMessages.Add(-1)
				if Tcp.IsConnectionClosed(err) {
					connection.Close()
					return
				}
				continue
			}
			connection.messageId++
			messageId := connection.messageId
			if infoLogger := connection.infoLogger; infoLogger != nil {
				infoLogger.Log("Received message #" + Helpers.Uint64ToString(messageId))
			}
			if connection.rateLimiterBytes != nil && !connection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
				connection.byteRateLimiterExceeded.Add(1)
				connection.unprocessedMessages.Add(-1)
				if connection.warningLogger != nil {
					connection.warningLogger.Log("Byte rate limiter exceeded for message #" + Helpers.Uint64ToString(messageId))
				}
				continue
			}
			if connection.rateLimiterMessages != nil && !connection.rateLimiterMessages.Consume(1) {
				connection.messageRateLimiterExceeded.Add(1)
				connection.unprocessedMessages.Add(-1)
				if connection.warningLogger != nil {
					connection.warningLogger.Log("Message rate limiter exceeded for message #" + Helpers.Uint64ToString(messageId))
				}
				continue
			}
			message, err := Message.Deserialize(messageBytes, connection.GetName())
			if err != nil {
				connection.invalidMessagesReceived.Add(1)
				connection.unprocessedMessages.Add(-1)
				if warningLogger := connection.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("failed to deserialize message #"+Helpers.Uint64ToString(messageId), err).Error())
				}
				continue
			}
			if err := connection.validateMessage(message); err != nil {
				connection.invalidMessagesReceived.Add(1)
				connection.unprocessedMessages.Add(-1)
				if warningLogger := connection.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("failed to validate message #"+Helpers.Uint64ToString(messageId), err).Error())
				}
				continue
			}
			if infoLogger := connection.infoLogger; infoLogger != nil {
				infoLogger.Log("queueing message #" + Helpers.Uint64ToString(messageId) + " for processing")
			}
			if message.IsResponse() {
				if err := connection.addSyncResponse(message); err != nil {
					connection.invalidSyncResponsesReceived.Add(1)
					if warningLogger := connection.warningLogger; warningLogger != nil {
						warningLogger.Log(Error.New("failed to add sync response for message #"+Helpers.Uint64ToString(messageId), err).Error())
					}
				} else {
					connection.validMessagesReceived.Add(1)
				}
				connection.unprocessedMessages.Add(-1)
				continue
			} else {
				connection.validMessagesReceived.Add(1)
				connection.processingChannel <- &messageInProcess{
					message: message,
					id:      messageId,
				}
			}
		}
	}
}

func (connection *TcpConnection) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := connection.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncToken()) > maxSyncTokenSize {
		return Error.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Error.New("Message missing topic", nil)
	}
	if maxTopicSize := connection.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Error.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := connection.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Error.New("Message payload exceeds maximum size", nil)
	}
	return nil
}
