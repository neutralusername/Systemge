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

	websocketConnection := server.NewWebsocketConnection(websocketConn)

	session, err := server.clientSessionManager.CreateSession("")
	if err != nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.CreateSessionFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
				Event.Identity:     "",
				Event.IdentityType: Event.WebsocketConnection,
				Event.Address:      websocketConnection.GetAddress(),
			},
		))
		websocketConn.Close()
		return nil
	}
	session.Set("connection", websocketConnection)

	websocketConnection.waitGroup.Add(1)
	server.waitGroup.Add(1)
	go server.websocketConnectionDisconnect(session, websocketConnection)
	go server.receptionRoutine(websocketConnection)

	return nil
}

func (server *WebsocketServer) websocketConnectionDisconnect(session *SessionManager.Session, websocketConnection *WebsocketConnection) {
	defer server.waitGroup.Done()

	select {
	case <-websocketConnection.stopChannel:
	case <-server.stopChannel:
		websocketConnection.Close()
	}

	websocketConnection.waitGroup.Wait()

	server.onEvent(Event.NewInfoNoOption(
		Event.HandlingDisconnection,
		"disconnecting websocketConnection",
		Event.Context{
			Event.Circumstance: Event.HandleDisconnection,
			Event.IdentityType: Event.WebsocketConnection,
			Event.Identity:     session.GetIdentity(),
			Event.Address:      websocketConnection.GetAddress(),
		},
	))
	server.removeWebsocketConnection(websocketConnection)

	server.onEvent(Event.NewInfoNoOption(
		Event.HandledDisconnection,
		"websocketConnection disconnected",
		Event.Context{
			Event.Circumstance: Event.HandleDisconnection,
			Event.IdentityType: Event.WebsocketConnection,
			Event.ClientId:     websocketConnection.GetId(),
			Event.Address:      websocketConnection.GetAddress(),
		},
	))
}
func (server *WebsocketServer) removeWebsocketConnection(websocketConnection *WebsocketConnection) {
	server.websocketConnectionMutex.Lock()
	defer server.websocketConnectionMutex.Unlock()
	delete(server.websocketConnections, websocketConnection.GetId())
	for groupId := range server.websocketConnectionGroups[websocketConnection.GetId()] {
		delete(server.websocketConnectionGroups[websocketConnection.GetId()], groupId)
		delete(server.groupsWebsocketConnections[groupId], websocketConnection.GetId())
		if len(server.groupsWebsocketConnections[groupId]) == 0 {
			delete(server.groupsWebsocketConnections, groupId)
		}
	}
}
