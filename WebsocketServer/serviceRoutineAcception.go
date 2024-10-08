package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) acceptionRoutine() {
	defer func() {
		if server.eventHandler != nil {
			server.onEvent(Event.New(
				Event.AcceptionRoutineEnds,
				Event.Context{},
				Event.Continue,
			))
		}
		server.waitGroup.Done()
	}()

	if server.eventHandler != nil {
		event := server.onEvent(Event.New(
			Event.AcceptionRoutineBegins,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	handleAcceptionWrapper := func(websocketClient *WebsocketClient.WebsocketClient) {
		if err := server.acceptionHandler(websocketClient); err != nil {
			websocketClient.Close()
			server.ClientsRejected.Add(1)
		} else {
			server.ClientsAccepted.Add(1)
		}
	}

	for {
		websocketClient, err := server.websocketListener.Accept(server.config.WebsocketClientConfig, server.config.AcceptTimeoutMs)
		if err != nil {
			websocketClient.Close()
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.AcceptClientFailed,
					Event.Context{
						Event.Address: websocketClient.GetAddress(),
					},
					Event.Skip,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					break
				}
			}
			continue
		}

		if server.config.HandleClientsSequentially {
			handleAcceptionWrapper(websocketClient)
		} else {
			server.waitGroup.Add(1)
			go func(websocketClient *WebsocketClient.WebsocketClient) {
				handleAcceptionWrapper(websocketClient)
				server.waitGroup.Done()
			}(websocketClient)
		}
	}
}

/*
	identity := ""
	if server.handshakeHandler != nil {
		identity_, err := server.handshakeHandler(websocketClient)
		if err != nil {
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.HandshakeFailed,
					Event.Context{
						Event.Address:  websocketClient.GetAddress(),
						Event.Identity: identity,
						Event.Error:    err.Error(),
					},
					Event.Skip,
					Event.Continue,
				))
				if event.GetAction() == Event.Skip {
					return err
				}
			} else {
				return err
			}
		}
		identity = identity_
	}
*/
