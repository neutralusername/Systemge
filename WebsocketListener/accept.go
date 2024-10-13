package WebsocketListener

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) accept(timeoutMs uint32) (*WebsocketClient.WebsocketClient, error) {
	var deadline <-chan time.Time
	if timeoutMs > 0 {
		deadline = time.After(time.Duration(timeoutMs) * time.Millisecond)
	}
	select {
	case <-listener.stopChannel:
		return nil, errors.New("listener stopped")
	case <-deadline:
		return nil, errors.New("accept timeout")
	case upgraderResponseChannel := <-listener.upgadeRequests:

		select {
		case <-listener.stopChannel:
			return nil, errors.New("listener stopped")
		case <-deadline:
			return nil, errors.New("accept timeout")
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
	listener.mutex.RLock()
	if listener.status != Status.Started {
		listener.mutex.RUnlock()
		return nil, errors.New("listener not started")
	}
	listener.waitgroup.Add(1)
	defer listener.waitgroup.Done()
	listener.mutex.RUnlock()

	return listener.accept()
}

func (listener *WebsocketListener) AcceptTimeout(timeoutMs uint32) (*WebsocketClient.WebsocketClient, error) {
	listener.mutex.RLock()
	if listener.status != Status.Started {
		listener.mutex.RUnlock()
		return nil, errors.New("listener not started")
	}
	listener.waitgroup.Add(1)
	defer listener.waitgroup.Done()
	listener.mutex.RUnlock()

	listener.SetAcceptDeadline(timeoutMs)
	return listener.accept()
}
