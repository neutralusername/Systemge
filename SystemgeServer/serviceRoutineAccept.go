package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) connectionRoutine() {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.ConnectionRoutineFinished,
			"stopped systemgeServer connection routine",
			Event.Context{
				Event.Circumstance: Event.ConnectionRoutine,
				Event.ClientType:   Event.SystemgeConnection,
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
			Event.ClientType:   Event.SystemgeConnection,
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

	if event := server.onEvent(Event.NewInfo(
		Event.HandlingConnection,
		"handling systemgeConnection connection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.HandleConnection,
			Event.ClientType:   Event.SystemgeConnection,
		},
	)); !event.IsInfo() {
		return event.GetError()
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
				Event.ClientType:   Event.SystemgeConnection,
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
				Event.ClientType:    Event.SystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetAddress(),
			},
		))
		connection.Close()
		return nil
	}
	session.Set("connection", connection)

	server.waitGroup.Add(1)
	go server.handleSystemgeDisconnect(connection)

	server.onEvent(Event.NewInfoNoOption(
		Event.HandledConnection,
		"systemgeConnection connected",
		Event.Context{
			Event.Circumstance:  Event.HandleConnection,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))

	return nil
}

func (server *SystemgeServer) handleSystemgeDisconnect(connection SystemgeConnection.SystemgeConnection) {
	defer server.waitGroup.Done()

	select {
	case <-connection.GetCloseChannel():
	case <-server.stopChannel:
		connection.Close()
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.HandlingDisconnection,
		"disconnecting systemgeConnection",
		Event.Context{
			Event.Circumstance:  Event.HandleDisconnection,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))

	server.removeSystemgeConnection(connection.GetName())

	server.onEvent(Event.NewInfoNoOption(
		Event.HandledDisconnection,
		"systemgeConnection disconnected",
		Event.Context{
			Event.Circumstance:  Event.HandleDisconnection,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))
}

func (server *SystemgeServer) removeSystemgeConnection(name string) {
	sessions := server.sessionManager.GetSessions(name)
	for _, session := range sessions {
		session.GetTimeout().Trigger()
	}
}
