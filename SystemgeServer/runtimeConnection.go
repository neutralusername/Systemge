package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) RemoveConnection(name string) error {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()
	if server.status != Status.Started {
		return Event.New("server is not started", nil)
	}
	if connection, ok := server.clients[name]; ok {
		connection.Close()
		return nil
	}
	return Event.New("connection not found", nil)
}

func (server *SystemgeServer) GetConnectionNamesAndAddresses() map[string]string {
	server.statusMutex.RLock()
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
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
	Server.mutex.Lock()
	defer func() {
		Server.mutex.Unlock()
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
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
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
	server.mutex.Lock()
	defer func() {
		server.mutex.Unlock()
		server.statusMutex.RUnlock()
	}()
	if server.status != Status.Started {
		return 0
	}
	return len(server.clients)
}
