package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) acceptRoutine() {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.AcceptionRoutineFinished,
			"stopped systemgeServer acception routine",
			Event.Context{
				Event.Circumstance: Event.AcceptionRoutine,
				Event.ClientType:   Event.SystemgeConnection,
			},
		))
		server.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.AcceptionRoutineStarted,
		"started systemgeServer acception routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptionRoutine,
			Event.ClientType:   Event.SystemgeConnection,
		},
	)); !event.IsInfo() {
		return
	}

	for err := server.acceptSystemgeConnection(); err == nil; {
	}
}

func (server *SystemgeServer) acceptSystemgeConnection() error {
	select {
	case <-server.stopChannel:
		return errors.New("systemgeServer stopped")
	default:
	}

	if event := server.onEvent(Event.NewInfo(
		Event.HandlingAcception,
		"accepting systemgeConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.HandleAcception,
			Event.ClientType:   Event.SystemgeConnection,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	connection, err := server.listener.AcceptConnection(server.config.TcpSystemgeConnectionConfig, server.eventHandler)
	if err != nil {
		event := server.onEvent(Event.NewInfo(
			Event.HandleAcceptionFailed,
			err.Error(),
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.HandleAcception,
				Event.ClientType:   Event.SystemgeConnection,
			},
		))
		if !event.IsInfo() {
			return event.GetError()
		} else {
			return nil
		}
	}

	server.mutex.Lock()
	if _, ok := server.clients[connection.GetName()]; ok {
		event := server.onEvent(Event.NewInfo(
			Event.DuplicateName,
			"duplicate name",
			Event.Cancel,
			Event.Cancel,
			Event.Skip,
			Event.Context{
				Event.Circumstance:  Event.HandleAcception,
				Event.ClientType:    Event.SystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetAddress(),
			},
		))
		server.mutex.Unlock()
		connection.Close()
		if !event.IsInfo() {
			return errors.New("duplicate name")
		} else {
			return nil
		}
	}
	server.clients[connection.GetName()] = nil
	server.mutex.Unlock()

	event := server.onEvent(Event.NewInfo(
		Event.HandledAcception,
		"systemgeConnection accepted",
		Event.Cancel,
		Event.Skip,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleAcception,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))
	if event.IsError() {
		connection.Close()
		server.removeSystemgeConnection(connection.GetName())
		return event.GetError()
	}
	if event.IsWarning() {
		connection.Close()
		server.removeSystemgeConnection(connection.GetName())
		return nil
	}

	server.mutex.Lock()
	server.clients[connection.GetName()] = connection
	server.mutex.Unlock()

	server.waitGroup.Add(1)
	go server.handleSystemgeDisconnect(connection)

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
			Event.Circumstance:  Event.Disconnection,
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
			Event.Circumstance:  Event.Disconnection,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))
}

func (server *SystemgeServer) removeSystemgeConnection(name string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.clients, name)
}
