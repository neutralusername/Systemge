package ListenerChannel

import (
	"errors"

	"github.com/neutralusername/Systemge/ConnectionChannel"
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *ChannelListener[T]) Accept(timeoutNs int64) (*ConnectionChannel.ChannelConnection[T], error) {

	timeout := Tools.NewTimeout(timeoutNs, nil, false)
	connection, err := listener.accept(timeout.GetIsExpiredChannel())
	timeout.Trigger()

	return connection, err
}

func (listener *ChannelListener[T]) accept(cancel <-chan struct{}) (*ConnectionChannel.ChannelConnection[T], error) {
	select {
	case <-listener.stopChannel:
		listener.ClientsFailed.Add(1)
		return nil, errors.New("listener stopped")

	case <-cancel:
		listener.ClientsFailed.Add(1)
		return nil, errors.New("accept canceled")

	case connectionRequest := <-listener.connectionChannel:
		listener.ClientsAccepted.Add(1)
		return ConnectionChannel.New(connectionRequest.SendToListener, connectionRequest.ReceiveFromListener), nil
	}
}
