package WebsocketListener

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) Accept(config *Config.WebsocketClient, timeoutMs uint32) (*WebsocketClient.WebsocketClient, error) {
	acceptRequest := &acceptRequest{
		upgraderResponseChannel: make(chan *upgraderResponse),
		timeoutMs:               timeoutMs,
		triggered:               sync.WaitGroup{},
	}
	acceptRequest.triggered.Add(1)
	listener.pool.AddItems(true, acceptRequest)

	var deadline <-chan time.Time
	if timeoutMs > 0 {
		deadline = time.After(time.Duration(timeoutMs) * time.Millisecond)
	}
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
		websocketClient, err := WebsocketClient.New(config, upgraderResponse.websocketConn)
		if err != nil {
			listener.ClientsFailed.Add(1)
			upgraderResponse.websocketConn.Close()
			return nil, err
		}

		listener.ClientsAccepted.Add(1)
		return websocketClient, nil
	}
}
