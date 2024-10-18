package connectionChannel

import (
	"errors"

	"github.com/neutralusername/systemge/tools"
)

func (client *ChannelConnection[T]) ReadChannel() <-chan T {
	return tools.ChannelCall(func() (T, error) {
		return client.Read(0)
	})
}

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
		case data := <-connection.receiveChannel:
			connection.MessagesReceived.Add(1)
			connection.readTimeout.Trigger()
			connection.readTimeout = nil
			return data, nil

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
