package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
)

func (connection *TcpConnection) UnprocessedMessagesCount() uint32 {
	return connection.processingChannelSemaphore.AvailableAcquires()
}

func (connection *TcpConnection) IsProcessingLoopRunning() bool {
	connection.processMutex.Lock()
	defer connection.processMutex.Unlock()
	return connection.processingLoopStopChannel != nil
}

func (connection *TcpConnection) GetNextMessage() (*Message.Message, error) {
	connection.processMutex.Lock()
	defer connection.processMutex.Unlock()
	if connection.processingLoopStopChannel != nil {
		return nil, Error.New("Processing loop already running", nil)
	}
	var timeout <-chan time.Time
	if connection.config.TcpReceiveTimeoutMs > 0 {
		timeout = time.After(time.Duration(connection.config.TcpReceiveTimeoutMs) * time.Millisecond)
	}

	select {
	case message := <-connection.processingChannel:
		if message == nil {
			return nil, Error.New("Connection closed and no remaining messages", nil)
		}
		connection.processingChannelSemaphore.ReleaseBlocking()
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Retrieved message \"" + Helpers.GetPointerId(message) + "\" in GetNextMessage()")
		}
		return message, nil
	case <-timeout:
		return nil, Error.New("Timeout while waiting for message", nil)
	}
}
