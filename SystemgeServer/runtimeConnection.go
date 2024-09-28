package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) RemoveConnection(name string) error {
	server.statusMutex.RLock()
	server.mutex.RLock()
	defer func() {
		server.mutex.RUnlock()
		server.statusMutex.RUnlock()
	}()
	if server.status != Status.Started {
		return errors.New("server is not started")
	}

	if event := server.onEvent(Event.NewInfo(
		Event.DisconnectingClient,
		"disconnecting systemgeConnection",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.DisconnectClientRuntime,
			Event.ClientType:   Event.SystemgeConnection,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	connection, ok := server.clients[name]
	if !ok {
		return errors.New("connection not found")
	}
	connection.Close()
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
