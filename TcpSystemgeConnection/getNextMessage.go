package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
)

func (connection *TcpSystemgeConnection) AvailableMessageCount() uint32 {
	return connection.messageChannelSemaphore.AvailableAcquires()
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
