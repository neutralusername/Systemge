package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) Accept(config *Config.WebsocketClient, eventHandler Event.Handler) (*WebsocketClient.WebsocketClient, error) {
	if event := listener.onEvent(Event.New(
		Event.AcceptingClient,
		Event.Context{},
		Event.Continue,
		Event.Cancel,
	)); event.GetAction() == Event.Cancel {
		return nil, errors.New("accept canceled")
	}

	websocketConn := <-listener.connectionChannel
	if websocketConn == nil {
		listener.onEvent(Event.New(
			Event.ReceivedNilValueFromChannel,
			Event.Context{},
			Event.Cancel,
		))
		return nil, errors.New("received nil value from connection channel")
	}

	websocketClient, err := WebsocketClient.New(config, websocketConn, eventHandler)
	if err != nil {
		listener.onEvent(Event.New(
			Event.CreateClientFailed,
			Event.Context{
				Event.Address: websocketConn.RemoteAddr().String(),
				Event.Error:   err.Error(),
			},
			Event.Cancel,
		))
		listener.ClientsFailed.Add(1)
		return nil, err
	}

	if event := listener.onEvent(Event.New(
		Event.AcceptedClient,
		Event.Context{
			Event.Address: websocketConn.RemoteAddr().String(),
		},
		Event.Continue,
		Event.Cancel,
	)); event.GetAction() == Event.Cancel {
		websocketClient.Close()
		listener.ClientsRejected.Add(1)
		return nil, errors.New("client rejected")
	}

	listener.ClientsAccepted.Add(1)
	return websocketClient, nil
}
