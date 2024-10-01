package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/SessionManager"
)

func (server *WebsocketServer) sessionRoutine() {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.SessionRoutineEnds,
			"stopped websocketConnection session routine",
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
				Event.IdentityType: Event.WebsocketConnection,
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
			Event.IdentityType: Event.WebsocketConnection,
		},
	)); !event.IsInfo() {
		return
	}

	for err := server.handleNewSession(); err == nil; {
	}

}

func (server *WebsocketServer) handleNewSession() error {
	websocketConn := <-server.connectionChannel
	if websocketConn == nil {
		server.onEvent(Event.NewInfoNoOption(
			Event.ReceivedNilValueFromChannel,
			"received nil from websocketConnection channel",
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
				Event.ChannelType:  Event.WebsocketConnection,
			},
		))
		return errors.New("received nil from websocketConnection channel")
	}

	session, err := server.clientSessionManager.CreateSession("", map[string]any{
		"connection": websocketConn,
	})
	if err != nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.CreateSessionFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
				Event.Identity:     "",
				Event.IdentityType: Event.WebsocketConnection,
				Event.Address:      websocketConn.LocalAddr().String(),
			},
		))
		server.websocketConnectionsRejected.Add(1)
		websocketConn.Close()
		return nil
	}
	websocketConnection := server.NewWebsocketConnection(session.GetId(), websocketConn)

	server.websocketConnectionsAccepted.Add(1)
	websocketConnection.waitGroup.Add(1)
	server.waitGroup.Add(1)
	go server.websocketConnectionDisconnect(session, websocketConnection)
	go server.receptionRoutine(websocketConnection)

	return nil
}

func (server *WebsocketServer) websocketConnectionDisconnect(session *SessionManager.Session, websocketConnection *WebsocketConnection) {
	select {
	case <-websocketConnection.stopChannel:
	case <-session.GetTimeout().GetTriggeredChannel():
	case <-server.stopChannel:
	}

	session.GetTimeout().Trigger()
	websocketConnection.Close()

	websocketConnection.waitGroup.Wait()
	server.waitGroup.Done()
}
