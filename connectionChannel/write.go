package connectionChannel

import (
	"errors"

	"github.com/neutralusername/systemge/tools"
)

func (connection *ChannelConnection[D]) Write(data D, timeoutNs int64) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()

	connection.writeTimeout = tools.NewTimeout(
		timeoutNs,
		nil,
		false,
	)

	for {
		select {
		case connection.sendChannel <- data:
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

func (connection *ChannelConnection[D]) SetWriteDeadline(timeoutNs int64) {
	if writeTimeout := connection.writeTimeout; writeTimeout != nil {
		writeTimeout.Refresh(timeoutNs)
	}
}
