package WebsocketListener

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) accept() (*WebsocketClient.WebsocketClient, error) {
	acceptRequest := &acceptRequest{
		upgraderResponseChannel: make(chan *upgraderResponse),
		triggered:               sync.WaitGroup{},
	}
	acceptRequest.triggered.Add(1)
	listener.pool.AddItems(true, acceptRequest)

	var deadline <-chan time.Time
	if timeoutMs > 0 {
		deadline = time.After(time.Duration(timeoutMs) * time.Millisecond)
	}

	// this feels needlessly complicated
	select {
	case <-listener.stopChannel:
		listener.pool.RemoveItems(true, acceptRequest)
		acceptRequest.triggered.Done()
		return nil, errors.New("listener stopped")

	case <-deadline:
		listener.pool.RemoveItems(true, acceptRequest)
		acceptRequest.triggered.Done()
		return nil, errors.New("timeout")

	case upgraderResponse := <-acceptRequest.upgraderResponseChannel:
		listener.pool.RemoveItems(true, acceptRequest)
		acceptRequest.triggered.Done()
		if upgraderResponse.err != nil {
			return nil, upgraderResponse.err
		}
		websocketClient, err := WebsocketClient.New(upgraderResponse.websocketConn)
		if err != nil {
			listener.ClientsFailed.Add(1)
			upgraderResponse.websocketConn.Close()
			return nil, err
		}
		listener.ClientsAccepted.Add(1)
		return websocketClient, nil
	}
}

func (listener *WebsocketListener) SetAcceptDeadline(timeoutMs uint32) {

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
