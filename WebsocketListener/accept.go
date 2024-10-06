package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) Accept(config *Config.WebsocketClient) (*WebsocketClient.WebsocketClient, error) {

	websocketConn := <-listener.connectionChannel
	if websocketConn == nil {
		return nil, errors.New("received nil value from connection channel")
	}

	websocketClient, err := WebsocketClient.New(config, websocketConn)
	if err != nil {
		listener.ClientsFailed.Add(1)
		return nil, err
	}

	listener.ClientsAccepted.Add(1)
	return websocketClient, nil
}
