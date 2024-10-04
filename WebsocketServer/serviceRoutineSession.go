package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) sessionRoutine() {
	defer func() {
		if server.eventHandler != nil {
			server.onEvent(Event.New(
				Event.SessionRoutineEnds,
				Event.Context{},
				Event.Continue,
				Event.Cancel,
			))
		}
		server.waitGroup.Done()
	}()

	if server.eventHandler != nil {
		event := server.onEvent(Event.New(
			Event.SessionRoutineBegins,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	for {

		websocketClient, err := server.websocketListener.AcceptClient(server.config.WebsocketClientConfig, server.eventHandler)
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

		if server.eventHandler != nil {
			event := server.onEvent(Event.New(
				Event.CreatingSession,
				Event.Context{
					Event.Address: websocketClient.GetAddress(),
				},
				Event.Continue,
				Event.Skip,
				Event.Cancel,
			))
			if event.GetAction() == Event.Cancel {
				break
			}
			if event.GetAction() == Event.Skip {
				continue
			}
		}

		session, err := server.sessionManager.CreateSession("", map[string]any{
			"websocketClient": websocketClient,
		})
		if err != nil {
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.CreateSessionFailed,
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

		if server.eventHandler != nil {
			event := server.onEvent(Event.New(
				Event.CreatedSession,
				Event.Context{
					Event.SessionId: session.GetId(),
					Event.Address:   websocketClient.GetAddress(),
				},
				Event.Continue,
				Event.Cancel,
			))
			if event.GetAction() == Event.Cancel {
				websocketClient.Close()
				session.GetTimeout().Trigger()
				break
			}
		}

		server.waitGroup.Add(1)
		go server.websocketClientDisconnect(session, websocketClient)
	}
}

func (server *WebsocketServer) onCreateSession(session *Tools.Session) error {
	if server.eventHandler != nil {
		websocketClient, ok := session.Get("websocketClient")
		if !ok {
			return errors.New("websocketClient not found")
		}
		event := server.onEvent(Event.New(
			Event.OnCreateSession,
			Event.Context{
				Event.SessionId: session.GetId(),
				Event.Address:   websocketClient.(*WebsocketClient.WebsocketClient).GetAddress(),
			},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return errors.New("session rejected")
		}
	}
	return nil
}

func (server *WebsocketServer) websocketClientDisconnect(session *Tools.Session, websocketClient *WebsocketClient.WebsocketClient) {
	select {
	case <-websocketClient.GetCloseChannel():
	case <-session.GetTimeout().GetTriggeredChannel():
	case <-server.stopChannel:
	}

	if server.eventHandler != nil {
		server.onEvent(Event.New(
			Event.OnDisconnect,
			Event.Context{
				Event.Identity: session.GetId(),
				Event.Address:  websocketClient.GetAddress(),
			},
			Event.Continue,
		))
	}

	session.GetTimeout().Trigger()
	websocketClient.Close()

	server.waitGroup.Done()
}
