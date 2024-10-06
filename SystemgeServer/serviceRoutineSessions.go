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
			Event.AcceptionRoutineEnds,
			"stopped systemgeServer session routine",
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
			},
		))
		server.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.AcceptionRoutineBegins,
		"started systemgeServer session routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.SessionRoutine,
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
			Event.AcceptTcpSystemgeListenerFailed,
			err.Error(),
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
			},
		))
		if !event.IsInfo() {
			return event.GetError()
		} else {
			return nil
		}
	}

	session, err := server.sessionManager.CreateSession(connection.GetName(), map[string]any{
		"connection": connection,
	})
	if err != nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.CreateSessionFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.SessionRoutine,
				Event.Identity:     connection.GetName(),
				Event.Address:      connection.GetAddress(),
			},
		))
		connection.Close()
		return nil
	}

	server.waitGroup.Add(1)
	go server.handleSystemgeDisconnect(session, connection)

	return nil
}

func (server *SystemgeServer) handleSystemgeDisconnect(session *SessionManager.Session, connection SystemgeConnection.SystemgeConnection) {
	defer server.waitGroup.Done()

	select {
	case <-connection.GetCloseChannel():
	case <-session.GetTimeout().GetTriggeredChannel():
	case <-server.stopChannel:
	}

	session.GetTimeout().Trigger()
	connection.Close()
}
