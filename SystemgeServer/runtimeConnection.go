package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) RemoveIdentity(identity string) error {
	server.statusMutex.RLock()
	defer server.statusMutex.RUnlock()

	if server.status != Status.Started {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"systemgeServer already stopped",
			Event.Context{
				Event.Circumstance: Event.DisconnectClientRuntime,
				Event.ClientType:   Event.SystemgeConnection,
			},
		))
		return errors.New("systemgeServer already stopped")
	}

	for _, session := range server.sessionManager.GetSessions(identity) {
		connection := session.GetConnection()
	}

	connection, ok := server.clients[identity]
	if !ok {
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientDoesNotExist,
			"systemgeConnection does not exist",
			Event.Context{
				Event.Circumstance: Event.DisconnectClientRuntime,
				Event.ClientType:   Event.SystemgeConnection,
			},
		))
		return errors.New("systemgeConnection not found")
	}
	if connection == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.ClientNotAccepted,
			"systemgeConnection not accepted",
			Event.Context{
				Event.Circumstance: Event.DisconnectClientRuntime,
				Event.ClientType:   Event.SystemgeConnection,
			},
		))
		return errors.New("systemgeConnection not accepted")
	}
	connection.Close()

	server.onEvent(Event.NewInfoNoOption(
		Event.HandledDisconnection,
		"systemgeConnection disconnected",
		Event.Context{
			Event.Circumstance: Event.DisconnectClientRuntime,
			Event.ClientType:   Event.SystemgeConnection,
			Event.ClientName:   identity,
		},
	))
	return nil
}

func (server *SystemgeServer) GetConnectionNamesAndAddresses() map[string]string {
	server.statusMutex.RLock()
	server.mutex.RLock()
	defer func() {
		server.mutex.RUnlock()
		server.statusMutex.RUnlock()
	}()
	if server.status != Status.Started {
		return nil
	}
	names := make(map[string]string, len(server.clients))
	for name, connection := range server.clients {
		names[name] = connection.GetAddress()
	}
	return names
}

func (Server *SystemgeServer) GetConnections() []SystemgeConnection.SystemgeConnection {
	Server.statusMutex.RLock()
	Server.mutex.RLock()
	defer func() {
		Server.mutex.RUnlock()
		Server.statusMutex.RUnlock()
	}()
	if Server.status != Status.Started {
		return nil
	}
	connections := make([]SystemgeConnection.SystemgeConnection, 0, len(Server.clients))
	for _, connection := range Server.clients {
		connections = append(connections, connection)
	}
	return connections
}

func (server *SystemgeServer) GetConnection(name string) SystemgeConnection.SystemgeConnection {
	server.statusMutex.RLock()
	server.mutex.RLock()
	defer func() {
		server.mutex.RUnlock()
		server.statusMutex.RUnlock()
	}()
	if server.status != Status.Started {
		return nil
	}
	if connection, ok := server.clients[name]; ok {
		return connection
	}
	return nil
}

func (server *SystemgeServer) GetConnectionCount() int {
	server.statusMutex.RLock()
	server.mutex.RLock()
	defer func() {
		server.mutex.RUnlock()
		server.statusMutex.RUnlock()
	}()
	if server.status != Status.Started {
		return 0
	}
	return len(server.clients)
}
