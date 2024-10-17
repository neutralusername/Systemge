package connectionChannel

import (
	"errors"

	"github.com/neutralusername/Systemge/tools"
)

func (connection *ChannelConnection[T]) Read(timeoutNs int64) (T, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	connection.readTimeout = tools.NewTimeout(
		timeoutNs,
		nil,
		false,
	)

	for {
		select {
		case item := <-connection.receiveChannel:
			connection.MessagesReceived.Add(1)
			connection.readTimeout.Trigger()
			connection.readTimeout = nil
			return item, nil

		case <-connection.readTimeout.GetIsExpiredChannel():
			connection.readTimeout = nil
			var nilValue T
			return nilValue, errors.New("timeout")
		}
	}
}

func (connection *ChannelConnection[T]) SetReadDeadline(timeoutNs int64) {
	if readTimeout := connection.readTimeout; readTimeout != nil {
		readTimeout.Refresh(timeoutNs)
	}
}
