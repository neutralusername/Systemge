package listenerChannel

import (
	"errors"

	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func (listener *ChannelListener[T]) Accept(timeoutNs int64) (systemge.Connection[T], error) {

	timeout := tools.NewTimeout(timeoutNs, nil, false)
	connection, err := listener.accept(timeout.GetIsExpiredChannel())
	timeout.Trigger()

	return connection, err
}

func (listener *ChannelListener[T]) accept(cancel <-chan struct{}) (systemge.Connection[T], error) {
	select {
	case <-listener.stopChannel:
		listener.ClientsFailed.Add(1)
		return nil, errors.New("listener stopped")

	case <-cancel:
		listener.ClientsFailed.Add(1)
		return nil, errors.New("accept canceled")

	case connectionRequest := <-listener.connectionChannel:
		listener.ClientsAccepted.Add(1)
		return connectionChannel.New(connectionRequest.SendToListener, connectionRequest.ReceiveFromListener), nil
	}
}
