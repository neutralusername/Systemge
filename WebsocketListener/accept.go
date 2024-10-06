package WebsocketListener

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) Accept(config *Config.WebsocketClient, timeoutMs uint32) (*WebsocketClient.WebsocketClient, error) {
	upgraderResponseChannel := make(chan *upgraderResponse)
	listener.acceptChannel <- upgraderResponseChannel
	upgraderResponse := <-upgraderResponseChannel

	if upgraderResponse.err != nil {
		return nil, upgraderResponse.err
	}

	websocketClient, err := WebsocketClient.New(config, upgraderResponse.websocketConn)
	if err != nil {
		listener.ClientsFailed.Add(1)
		return nil, err
	}

	listener.ClientsAccepted.Add(1)
	return websocketClient, nil
}
