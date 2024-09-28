package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) acceptRoutine(stopChannel chan bool) {
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
	server.waitGroup.Add(1)

	select {
	case <-server.stopChannel:
		server.waitGroup.Done()
		return errors.New("systemgeServer stopped")
	default:
	}

	if event := server.onEvent(Event.NewInfo(
		Event.AcceptingClient,
		"accepting systemgeConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.AcceptionRoutine,
			Event.ClientType:   Event.SystemgeConnection,
		},
	)); !event.IsInfo() {
		server.waitGroup.Done()
		return event.GetError()
	}

	connection, err := server.listener.AcceptConnection(server.config.TcpSystemgeConnectionConfig, server.eventHandler)
	if err != nil {
		event := server.onEvent(Event.NewInfo(
			Event.AcceptingClientFailed,
			err.Error(),
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.AcceptionRoutine,
				Event.ClientType:   Event.SystemgeConnection,
			},
		))
		server.waitGroup.Done()
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
				Event.Circumstance:  Event.AcceptionRoutine,
				Event.ClientType:    Event.SystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetAddress(),
			},
		))
		server.mutex.Unlock()
		connection.Close()
		server.waitGroup.Done()
		if !event.IsInfo() {
			return errors.New("duplicate name")
		} else {
			return nil
		}
	}
	server.clients[connection.GetName()] = nil
	server.mutex.Unlock()

	event := server.onEvent(Event.NewInfo(
		Event.AcceptedClient,
		err.Error(),
		Event.Cancel,
		Event.Skip,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.AcceptionRoutine,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))
	if event.IsError() {
		connection.Close()
		server.waitGroup.Done()
		server.mutex.Lock()
		delete(server.clients, connection.GetName())
		server.mutex.Unlock()
		return event.GetError()
	}
	if event.IsWarning() {
		connection.Close()
		server.waitGroup.Done()
		server.mutex.Lock()
		delete(server.clients, connection.GetName())
		server.mutex.Unlock()
		return nil
	}

	server.mutex.Lock()
	server.clients[connection.GetName()] = connection
	server.mutex.Unlock()

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
		Event.DisconnectingClient,
		"disconnecting systemgeConnection",
		Event.Context{
			Event.Circumstance:  Event.Disconnection,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))

	server.mutex.Lock()
	delete(server.clients, connection.GetName())
	server.mutex.Unlock()

	server.onEvent(Event.NewInfoNoOption(
		Event.DisconnectedClient,
		"systemgeConnection disconnected",
		Event.Context{
			Event.Circumstance:  Event.Disconnection,
			Event.ClientType:    Event.SystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetAddress(),
		},
	))
}
