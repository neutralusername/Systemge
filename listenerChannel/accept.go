package listenerChannel

import (
	"errors"

	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func (listener *ChannelListener[D]) Accept(timeoutNs int64) (systemge.Connection[D], error) {
	listener.mutex.Lock()
	defer listener.mutex.Unlock()

	listener.timeout = tools.NewTimeout(timeoutNs, nil, false)
	defer func() {
		listener.timeout.Trigger()
		listener.timeout = nil
	}()
	select {
	case <-listener.stopChannel:
		listener.ClientsFailed.Add(1)
		return nil, errors.New("listener stopped")

	case <-listener.timeout.GetIsExpiredChannel():
		listener.ClientsFailed.Add(1)
		return nil, errors.New("accept canceled")

	case connectionRequest := <-listener.connectionChannel:
		listener.ClientsAccepted.Add(1)
		return connectionChannel.New(connectionRequest.SendToListener, connectionRequest.ReceiveFromListener, listener.connectionLifetimeNs), nil
	}
}

func (listener *ChannelListener[D]) SetAcceptDeadline(timeoutNs int64) {
	if timeout := listener.timeout; timeout != nil {
		timeout.Refresh(timeoutNs)
	}
}
