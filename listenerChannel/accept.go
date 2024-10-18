package listenerChannel

import (
	"errors"

	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func (listener *ChannelListener[D]) Accept(timeoutNs int64) (systemge.Connection[D], error) {

	timeout := tools.NewTimeout(timeoutNs, nil, false)
	select {
	case <-listener.stopChannel:
		listener.ClientsFailed.Add(1)
		return nil, errors.New("listener stopped")

	case <-timeout.GetIsExpiredChannel():
		listener.ClientsFailed.Add(1)
		return nil, errors.New("accept canceled")

	case connectionRequest := <-listener.connectionChannel:
		listener.ClientsAccepted.Add(1)
		return connectionChannel.New(connectionRequest.SendToListener, connectionRequest.ReceiveFromListener), nil
	}
}

func (listener *ChannelListener[D]) SetAcceptDeadline(timeoutNs int64) {

}
