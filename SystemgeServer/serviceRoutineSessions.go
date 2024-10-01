package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/SessionManager"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) sessionRoutine() {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.SessionRoutineEnds,
			"stopped systemgeServer session routine",
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
				Event.IdentityType: Event.SystemgeConnection,
			},
		))
		server.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.SessionRoutineBegins,
		"started systemgeServer session routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.SessionRoutine,
			Event.IdentityType: Event.SystemgeConnection,
		},
	)); !event.IsInfo() {
		return
	}

	for err := server.handleNewSession(); err == nil; {
	}
}

func (server *SystemgeServer) handleNewSession() error {
	select {
	case <-server.stopChannel:
		return errors.New("systemgeServer stopped")
	default:
	}

	connection, err := server.listener.AcceptConnection(server.config.TcpSystemgeConnectionConfig, server.eventHandler)
	if err != nil {
		event := server.onEvent(Event.NewInfo(
			Event.TcpSystemgeListenerAcceptFailed,
			err.Error(),
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
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
				Event.Circumstance: Event.SessionRoutine,
				Event.Identity:     connection.GetName(),
				Event.IdentityType: Event.SystemgeConnection,
				Event.Address:      connection.GetAddress(),
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
