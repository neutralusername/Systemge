package WebsocketListener

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) accept(cancel chan struct{}) (*WebsocketClient.WebsocketClient, error) {
	select {
	case <-listener.stopChannel:
		return nil, errors.New("listener stopped")

	case <-cancel:
		return nil, errors.New("accept canceled")

	case upgraderResponseChannel := <-listener.upgadeRequests:
		select {
		case <-listener.stopChannel:
			return nil, errors.New("listener stopped")

		case <-cancel:
			return nil, errors.New("accept canceled")

		case upgraderResponse := <-upgraderResponseChannel:
			if upgraderResponse.err != nil {
				return nil, upgraderResponse.err
			}
			websocketClient, err := WebsocketClient.New(upgraderResponse.websocketConn)
			if err != nil {
				upgraderResponse.websocketConn.Close()
				return nil, err
			}
			return websocketClient, nil
		}
	}
}

func (listener *WebsocketListener) Accept() (*WebsocketClient.WebsocketClient, error) {
	return listener.accept(make(chan struct{}))
}

func (listener *WebsocketListener) AcceptTimeout(timeoutMs uint32) (*WebsocketClient.WebsocketClient, error) {
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
