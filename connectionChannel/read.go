package connectionChannel

import (
	"errors"

	"github.com/neutralusername/systemge/tools"
)

func (connection *ChannelConnection[D]) Read(timeoutNs int64) (D, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	connection.readTimeout = tools.NewTimeout(
		timeoutNs,
		nil,
		false,
	)

	for {
		select {
		case data := <-connection.receiveChannel:
			connection.MessagesReceived.Add(1)
			connection.readTimeout.Trigger()
			connection.readTimeout = nil
			return data, nil

		case <-connection.readTimeout.GetIsExpiredChannel():
			connection.readTimeout = nil
			var nilValue D
			return nilValue, errors.New("timeout")
		}
	}
}

func (connection *ChannelConnection[D]) SetReadDeadline(timeoutNs int64) {
	if readTimeout := connection.readTimeout; readTimeout != nil {
		readTimeout.Refresh(timeoutNs)
	}
}
