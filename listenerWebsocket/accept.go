package listenerWebsocket

import (
	"errors"

	"github.com/neutralusername/systemge/connectionWebsocket"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func (listener *WebsocketListener) Accept(timeoutNs int64) (systemge.Connection[[]byte], error) {
	listener.mutex.Lock()
	defer listener.mutex.Unlock()

	listener.timeout = tools.NewTimeout(timeoutNs, nil, false)
	defer func() {
		listener.timeout.Trigger()
		listener.timeout = nil
	}()
	select {
	case <-listener.stopChannel:
		return nil, errors.New("listener stopped")

	case <-listener.timeout.GetIsExpiredChannel():
		return nil, errors.New("accept canceled")

	case upgraderResponseChannel := <-listener.upgradeRequests:
		select {
		case <-listener.stopChannel:
			return nil, errors.New("listener stopped")

		case <-listener.timeout.GetIsExpiredChannel():
			return nil, errors.New("accept canceled")

		case upgraderResponse := <-upgraderResponseChannel:
			if upgraderResponse.err != nil {
				listener.ClientsFailed.Add(1)
				return nil, upgraderResponse.err
			}
			websocketClient, err := connectionWebsocket.New(upgraderResponse.websocketConn, listener.incomingMessageByteLimit, listener.connectionLifetimeNs)
			if err != nil {
				listener.ClientsFailed.Add(1)
				upgraderResponse.websocketConn.Close()
				return nil, err
			}
			listener.ClientsAccepted.Add(1)
			return websocketClient, nil
		}
	}
}

func (listener *WebsocketListener) SetAcceptDeadline(timeoutNs int64) {
	if timeout := listener.timeout; timeout != nil {
		timeout.Refresh(timeoutNs)
	}
}
