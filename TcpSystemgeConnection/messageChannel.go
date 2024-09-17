package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
)

func (connection *TcpSystemgeConnection) StartMessageHandlingLoop(messageHandlingStopChannel chan<- bool) error {

}

func (connection *TcpSystemgeConnection) GetNextMessage() (*Message.Message, error) {
	connection.messageMutex.Lock()
	defer connection.messageMutex.Unlock()
	var timeout <-chan time.Time
	if connection.config.TcpReceiveTimeoutMs > 0 {
		timeout = time.After(time.Duration(connection.config.TcpReceiveTimeoutMs) * time.Millisecond)
	}

	select {
	case message := <-connection.messageChannel:
		if message == nil {
			return nil, Error.New("Connection closed and no remaining messages", nil)
		}
		connection.messageChannelSemaphore.ReleaseBlocking()
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Retrieved message \"" + Helpers.GetPointerId(message) + "\" in GetNextMessage()")
		}
		return message, nil
	case <-timeout:
		return nil, Error.New("Timeout while waiting for message", nil)
	}
}

func (connection *TcpSystemgeConnection) AvailableMessageCount() uint32 {
	return connection.messageChannelSemaphore.AvailableAcquires()
}

func (connection *TcpSystemgeConnection) addMessageToChannel(messageBytes []byte) error {
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
