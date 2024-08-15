package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Error"
)

func (server *SystemgeServer) GetClientConnectionsList() map[string]string { // map == name:address
	server.clientConnectionMutex.RLock()
	defer server.clientConnectionMutex.RUnlock()
	connections := make(map[string]string, len(server.clientConnections))
	for name, clientConnection := range server.clientConnections {
		connections[name] = clientConnection.netConn.RemoteAddr().String()
	}
	return connections
}

func (server *SystemgeServer) DisconnectClientConnection(name string) error {
	server.clientConnectionMutex.Lock()
	defer server.clientConnectionMutex.Unlock()
	if server.clientConnections[name] == nil {
		return Error.New("Connection with name \""+name+"\" does not exist", nil)
	}
	server.clientConnections[name].netConn.Close()
	return nil
}

// AddToSystemgeBlacklist adds an address to the systemge blacklist.
func (server *SystemgeServer) AddToSystemgeBlacklist(address string) {
	server.tcpServer.GetBlacklist().Add(address)
}

// RemoveFromSystemgeBlacklist removes an address from the systemge blacklist.
func (server *SystemgeServer) RemoveFromSystemgeBlacklist(address string) {
	server.tcpServer.GetBlacklist().Remove(address)
}

// GetSystemgeBlacklist returns a slice of addresses in the systemge blacklist.
func (server *SystemgeServer) GetSystemgeBlacklist() []string {
	return server.tcpServer.GetBlacklist().GetElements()
}

// AddToSystemgeWhitelist adds an address to the systemge whitelist.
func (server *SystemgeServer) AddToSystemgeWhitelist(address string) {
	server.tcpServer.GetWhitelist().Add(address)
}

// RemoveFromSystemgeWhitelist removes an address from the systemge whitelist.
func (server *SystemgeServer) RemoveFromSystemgeWhitelist(address string) {
	server.tcpServer.GetWhitelist().Remove(address)
}

// GetSystemgeWhitelist returns a slice of addresses in the systemge whitelist.
func (server *SystemgeServer) GetSystemgeWhitelist() []string {
	return server.tcpServer.GetWhitelist().GetElements()
}
