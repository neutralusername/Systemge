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
		mutex:                   sync.Mutex{},
		timedOut:                false,
	}
	listener.acceptChannel <- acceptRequest

	var deadline <-chan time.Time
	if timeoutMs > 0 {
		deadline = time.After(time.Duration(timeoutMs) * time.Millisecond)
	}
	select {
	case <-listener.stopChannel:
		acceptRequest.mutex.Lock()
		acceptRequest.timedOut = true
		acceptRequest.mutex.Unlock()
		return nil, errors.New("listener stopped")
	case <-deadline:
		acceptRequest.mutex.Lock()
		acceptRequest.timedOut = true
		acceptRequest.mutex.Unlock()
		return nil, errors.New("timeout")
	case upgraderResponse := <-acceptRequest.upgraderResponseChannel:
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
