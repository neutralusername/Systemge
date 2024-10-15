package ConnectionChannel

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

func (connection *ChannelConnection[T]) Write(messageBytes T, timeoutNs int64) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()

	connection.writeTimeout = Tools.NewTimeout(
		timeoutNs,
		nil,
		false,
	)

	for {
		select {
		case connection.sendChannel <- messageBytes:
			connection.MessagesSent.Add(1)
			connection.writeTimeout.Trigger()
			connection.writeTimeout = nil
			return nil

		case <-connection.writeTimeout.GetIsExpiredChannel():
			connection.writeTimeout = nil
			return errors.New("timeout")
		}
	}
}

func (connection *ChannelConnection[T]) SetWriteDeadline(timeoutNs int64) {
	if writeTimeout := connection.writeTimeout; writeTimeout != nil {
		writeTimeout.Refresh(timeoutNs)
	}
}
