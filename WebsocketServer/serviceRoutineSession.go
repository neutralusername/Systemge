package WebsocketServer

import (
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
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
			},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	for {
		if server.eventHandler != nil {
			event := server.onEvent(Event.New(
				Event.CreatingSession,
				Event.SessionRoutine,
				Event.Context{},
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

		session, err := server.sessionManager.CreateSession("", map[string]any{})
		if err != nil {
			if server.eventHandler != nil {
				if event := server.onEvent(Event.NewWarning(
					Event.CreateSessionFailed,
					err.Error(),
					Event.Cancel,
					Event.Skip,
					Event.Skip,
					Event.Context{
						Event.Circumstance: Event.SessionRoutine,
					},
				)); event.IsError() {
					break
				}
			}
			continue
		}

		if server.eventHandler != nil {
			server.onEvent(Event.NewInfoNoOption(
				Event.CreatedSession,
				"created session",
				Event.Context{
					Event.Circumstance: Event.SessionRoutine,
					Event.Identity:     session.GetId(),
				},
			))
		}
	}
}

func (server *WebsocketServer) onCreate(session *Tools.Session) error {

	websocketConnection, err := server.websocketListener.AcceptClient(session.GetId(), server.config.WebsocketClientConfig, server.eventHandler)
	if err != nil {
		if server.eventHandler != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.AcceptClientFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance: Event.OnCreateSession,
					Event.Identity:     session.GetId(),
					Event.Address:      websocketConnection.GetAddress(),
				},
			))
		}
		websocketConnection.Close()
		return err
	}

	session.Set("websocketConnection", websocketConnection)
	server.waitGroup.Add(1)
	go server.websocketConnectionDisconnect(session, websocketConnection)

	return nil
}

func (server *WebsocketServer) websocketConnectionDisconnect(session *Tools.Session, websocketConnection *WebsocketClient.WebsocketClient) {
	select {
	case <-websocketConnection.GetCloseChannel():
	case <-session.GetTimeout().GetTriggeredChannel():
	case <-server.stopChannel:
	}

	session.GetTimeout().Trigger()
	websocketConnection.Close()

	server.waitGroup.Done()
}
