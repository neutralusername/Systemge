package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/SessionManager"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) connectionRoutine() {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.ConnectionRoutineFinished,
			"stopped systemgeServer connection routine",
			Event.Context{
				Event.Circumstance: Event.ConnectionRoutine,
				Event.IdentityType: Event.SystemgeConnection,
			},
		))
		server.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.ConnectionRoutineBegins,
		"started systemgeServer connection routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ConnectionRoutine,
			Event.IdentityType: Event.SystemgeConnection,
		},
	)); !event.IsInfo() {
		return
	}

	for err := server.handleConnection(); err == nil; {
	}
}

func (server *SystemgeServer) handleConnection() error {
	select {
	case <-server.stopChannel:
		return errors.New("systemgeServer stopped")
	default:
	}

	connection, err := server.listener.AcceptConnection(server.config.TcpSystemgeConnectionConfig, server.eventHandler)
	if err != nil {
		event := server.onEvent(Event.NewInfo(
			Event.HandleConnectionFailed,
			err.Error(),
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.HandleConnection,
				Event.IdentityType: Event.SystemgeConnection,
			},
		))
		if !event.IsInfo() {
			return event.GetError()
		} else {
			return nil
		}
	}

	session, err := server.sessionManager.CreateSession(connection.GetName())
	if err != nil {
		server.onEvent(Event.NewError(
			Event.CreateSessionFailed,
			err.Error(),
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:  Event.HandleConnection,
				Event.IdentityType:  Event.SystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetAddress(),
			},
		))
		connection.Close()
		return nil
	}
	session.Set("connection", connection)

	server.waitGroup.Add(1)
	go server.handleSystemgeDisconnect(session, connection)

	return nil
}

func (server *SystemgeServer) handleSystemgeDisconnect(session *SessionManager.Session, connection SystemgeConnection.SystemgeConnection) {
	defer server.waitGroup.Done()

	select {
	case <-connection.GetCloseChannel():
	case <-server.stopChannel:
		connection.Close()
	}

	session.GetTimeout().Trigger()
	connection.Close()
}
