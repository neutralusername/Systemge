package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (server *WebsocketServer) sessionRoutine() {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.SessionRoutineEnds,
			"stopped websocketConnection session routine",
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
			},
		))
		server.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.SessionRoutineBegins,
		"started websocketConnection session routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.SessionRoutine,
		},
	)); !event.IsInfo() {
		return
	}

	for {
		session, err := server.sessionManager.CreateSession("", map[string]any{})
		websocketConnection, err := server.websocketListener.AcceptClient(session.GetId(), server.config.WebsocketClientConfig, server.eventHandler)
		if err != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.CreateSessionFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance: Event.SessionRoutine,
					Event.Identity:     session.GetId(),
					Event.Address:      websocketConnection.GetAddress(),
				},
			))
			websocketConnection.Close()
			continue
		}
		session.Set("websocketConnection", websocketConnection)

		server.waitGroup.Add(1)
		go server.websocketConnectionDisconnect(session, websocketConnection)
	}
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
