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
				Event.SessionRoutine,
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
			Event.SessionRoutine,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	for {

		websocketConnection, err := server.websocketListener.AcceptClient(server.config.WebsocketClientConfig, server.eventHandler)
		if err != nil {
			websocketConnection.Close()
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.AcceptClientFailed,
					Event.SessionRoutine,
					Event.Context{
						Event.Address: websocketConnection.GetAddress(),
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
				Event.SessionRoutine,
				Event.Context{
					Event.Address: websocketConnection.GetAddress(),
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
			"websocketConnection": websocketConnection,
		})
		if err != nil {
			if server.eventHandler != nil {
				event := server.onEvent(Event.New(
					Event.CreateSessionFailed,
					Event.SessionRoutine,
					Event.Context{
						Event.Address: websocketConnection.GetAddress(),
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
				Event.SessionRoutine,
				Event.Context{
					Event.SessionId: session.GetId(),
					Event.Address:   websocketConnection.GetAddress(),
				},
				Event.Continue,
				Event.Cancel,
			))
			if event.GetAction() == Event.Cancel {
				websocketConnection.Close()
				session.GetTimeout().Trigger()
				break
			}
		}

		server.waitGroup.Add(1)
		go server.websocketConnectionDisconnect(session, websocketConnection)
	}
}

func (server *WebsocketServer) onCreateSession(session *Tools.Session) error {
	if server.eventHandler != nil {
		websocketConnection, ok := session.Get("websocketConnection")
		if !ok {
			return errors.New("websocketConnection not found")
		}
		event := server.onEvent(Event.New(
			Event.OnCreateSession,
			Event.SessionRoutine,
			Event.Context{
				Event.SessionId: session.GetId(),
				Event.Address:   websocketConnection.(*WebsocketClient.WebsocketClient).GetAddress(),
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

func (server *WebsocketServer) websocketConnectionDisconnect(session *Tools.Session, websocketConnection *WebsocketClient.WebsocketClient) {
	select {
	case <-websocketConnection.GetCloseChannel():
	case <-session.GetTimeout().GetTriggeredChannel():
	case <-server.stopChannel:
	}

	if server.eventHandler != nil {
		server.onEvent(Event.New(
			Event.OnDisconnect,
			Event.SessionRoutine,
			Event.Context{
				Event.Identity: session.GetId(),
				Event.Address:  websocketConnection.GetAddress(),
			},
			Event.Continue,
		))
	}

	session.GetTimeout().Trigger()
	websocketConnection.Close()

	server.waitGroup.Done()
}
