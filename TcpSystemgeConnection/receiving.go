package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpConnection) addMessageToProcessingChannelLoop() {
	if connection.infoLogger != nil {
		connection.infoLogger.Log("Started receiving messages")
	}
	for {
		select {
		case <-connection.closeChannel:
			if connection.infoLogger != nil {
				connection.infoLogger.Log("Stopped receiving messages")
			}
			return
		default:
			if err := connection.addMessageToProcessingChannel(); err != nil {
				if connection.warningLogger != nil {
					connection.warningLogger.Log(Error.New("failed to add message to processing channel", err).Error())
				}
			}
		}
	}
}

func (connection *TcpConnection) addMessageToProcessingChannel() error {
	messageBytes, err := connection.receive()
	if err != nil {
		if Tcp.IsConnectionClosed(err) {
			connection.Close()
		}
		return Error.New("failed to receive message", err)
	}
	if connection.rateLimiterBytes != nil && !connection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		connection.byteRateLimiterExceeded.Add(1)
		return Error.New("byte rate limiter exceeded", nil)
	}
	if connection.rateLimiterMessages != nil && !connection.rateLimiterMessages.Consume(1) {
		connection.messageRateLimiterExceeded.Add(1)
		return Error.New("message rate limiter exceeded", nil)
	}
	message, err := Message.Deserialize(messageBytes, connection.GetName())
	if err != nil {
		connection.invalidMessagesReceived.Add(1)
		return Error.New("failed to deserialize message", err)
	}
	if err := connection.validateMessage(message); err != nil {
		connection.invalidMessagesReceived.Add(1)
		return Error.New("failed to validate message", err)
	}
	if message.IsResponse() {
		if err := connection.addSyncResponse(message); err != nil {
			connection.invalidSyncResponsesReceived.Add(1)
			return Error.New("failed to add sync response", err)
		} else {
			connection.validMessagesReceived.Add(1)
		}
		return nil
	} else {
		connection.processingChannelSemaphore.AcquireBlocking()
		connection.closedMutex.Lock()
		defer connection.closedMutex.Unlock()
		if connection.closed {
			connection.processingChannelSemaphore.ReleaseBlocking()
			return Error.New("connection closed before message could be added to processing channel", nil)
		}
		connection.validMessagesReceived.Add(1)
		connection.processingChannel <- message
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Added message \"" + Helpers.GetPointerId(message) + "\" to processing channel")
		}
		return nil
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
