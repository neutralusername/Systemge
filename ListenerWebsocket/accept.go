package ListenerWebsocket

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/ConnectionWebsocket"
)

func (listener *WebsocketListener) Accept(timeoutMs uint32) (*ConnectionWebsocket.WebsocketConnection, error) {
	var deadline <-chan time.Time = time.After(time.Duration(timeoutMs) * time.Millisecond)
	var cancel chan struct{} = make(chan struct{})
	go func() {
		select {
		case <-deadline:
			close(cancel)
		case <-cancel:
			close(cancel)
		}
	}()
	websocketConnection, err := listener.accept(cancel)
	close(cancel)
	return websocketConnection, err
}

func (listener *WebsocketListener) accept(cancel <-chan struct{}) (*ConnectionWebsocket.WebsocketConnection, error) {
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
