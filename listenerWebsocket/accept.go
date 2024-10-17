package listenerWebsocket

import (
	"errors"

	"github.com/neutralusername/Systemge/connectionWebsocket"
	"github.com/neutralusername/Systemge/systemge"
	"github.com/neutralusername/Systemge/tools"
)

func (listener *WebsocketListener) Accept(timeoutNs int64) (systemge.Connection[[]byte], error) {
	timeout := tools.NewTimeout(timeoutNs, nil, false)
	connection, err := listener.accept(timeout.GetIsExpiredChannel())
	timeout.Trigger()

	return connection, err
}

func (listener *WebsocketListener) accept(cancel <-chan struct{}) (systemge.Connection[[]byte], error) {
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
			websocketClient, err := connectionWebsocket.New(upgraderResponse.websocketConn)
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
