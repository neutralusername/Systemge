package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpSystemgeConnection) receiveMessageLoop() {
	if connection.infoLogger != nil {
		connection.infoLogger.Log("Started receiving messages")
	}
	for {
		select {
		case <-connection.closeChannel:
			if connection.infoLogger != nil {
				connection.infoLogger.Log("Stopped receiving messages")
			}
			close(connection.receiveLoopStopChannel)
			return
		default:
			select {
			case <-connection.closeChannel:
				if connection.infoLogger != nil {
					connection.infoLogger.Log("Stopped receiving messages")
				}
				close(connection.receiveLoopStopChannel)
				return
			case <-connection.messageChannelSemaphore.GetChannel():
				messageBytes, bytesReceived, err := connection.messageReceiver.ReceiveNextMessage()
				connection.bytesReceived.Add(uint64(bytesReceived))
				if err != nil {
					if Tcp.IsConnectionClosed(err) {
						close(connection.receiveLoopStopChannel)
						connection.Close()
						return
					}
					if connection.warningLogger != nil {
						connection.warningLogger.Log(Event.New("failed to receive message", err).Error())
					}
					continue
				}
				if err := connection.addMessageToChannel(messageBytes); err != nil {
					if connection.warningLogger != nil {
						connection.warningLogger.Log(Event.New("failed to add message to processing channel", err).Error())
					}
					connection.messageChannelSemaphore.ReleaseBlocking()
				}
			}

		}
	}
}

func (connection *TcpSystemgeConnection) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := connection.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncToken()) > maxSyncTokenSize {
		return Event.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Event.New("Message missing topic", nil)
	}
	if maxTopicSize := connection.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Event.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := connection.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Event.New("Message payload exceeds maximum size", nil)
	}
	return nil
}

func (connection *TcpSystemgeConnection) addMessageToChannel(messageBytes []byte) error {
	if connection.rateLimiterBytes != nil && !connection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		connection.byteRateLimiterExceeded.Add(1)
		return Event.New("byte rate limiter exceeded", nil)
	}
	if connection.rateLimiterMessages != nil && !connection.rateLimiterMessages.Consume(1) {
		connection.messageRateLimiterExceeded.Add(1)
		return Event.New("message rate limiter exceeded", nil)
	}
	message, err := Message.Deserialize(messageBytes, connection.GetName())
	if err != nil {
		connection.invalidMessagesReceived.Add(1)
		return Event.New("failed to deserialize message", err)
	}
	if err := connection.validateMessage(message); err != nil {
		connection.invalidMessagesReceived.Add(1)
		return Event.New("failed to validate message", err)
	}
	if message.IsResponse() {
		if err := connection.addSyncResponse(message); err != nil {
			connection.invalidSyncResponsesReceived.Add(1)
			return Event.New("failed to add sync response", err)
		}
		connection.validMessagesReceived.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return nil
	} else {
		connection.validMessagesReceived.Add(1)
		connection.messageChannel <- message
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Added message \"" + Helpers.GetPointerId(message) + "\" to processing channel")
		}
		return nil
	}
}
