package ListenerWebsocket

import (
	"errors"

	"github.com/neutralusername/Systemge/ConnectionWebsocket"
	"github.com/neutralusername/Systemge/Systemge"
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *WebsocketListener) Accept(timeoutNs int64) (Systemge.Connection[[]byte], error) {
	timeout := Tools.NewTimeout(timeoutNs, nil, false)
	connection, err := listener.accept(timeout.GetIsExpiredChannel())
	timeout.Trigger()

	return connection, err
}

func (listener *WebsocketListener) accept(cancel <-chan struct{}) (Systemge.Connection[[]byte], error) {
	select {
	case <-listener.stopChannel:
		return nil, errors.New("listener stopped")

	case <-cancel:
		return nil, errors.New("accept canceled")

	case upgraderResponseChannel := <-listener.upgradeRequests:
		select {
		case <-listener.stopChannel:
			return nil, errors.New("listener stopped")

		case <-cancel:
			return nil, errors.New("accept canceled")

		case upgraderResponse := <-upgraderResponseChannel:
			if upgraderResponse.err != nil {
				listener.ClientsFailed.Add(1)
				return nil, upgraderResponse.err
			}
			websocketClient, err := ConnectionWebsocket.New(upgraderResponse.websocketConn)
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
