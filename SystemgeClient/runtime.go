package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Config"
)

// Returns a slice of addresses of servers that this client is connected to.
func (client *SystemgeClient) GetServerConnectionsList() []string {
	client.serverConnectionMutex.RLock()
	defer client.serverConnectionMutex.RUnlock()
	connections := make([]string, len(client.serverConnections))
	i := 0
	for address := range client.serverConnections {
		connections[i] = address
		i++
	}
	return connections
}

// Returns a slice of addresses of servers that this client is attempting to connect to.
func (client *SystemgeClient) GetServerConnectionAttemptsList() []string {
	client.serverConnectionMutex.RLock()
	defer client.serverConnectionMutex.RUnlock()
	attempts := make([]string, len(client.serverConnectionAttempts))
	i := 0
	for address := range client.serverConnectionAttempts {
		attempts[i] = address
		i++
	}
	return attempts
}

// Connects to a Systemge Server
// This connection is used to send async and sync requests and receive sync responses for their corresponding requests.
func (client *SystemgeClient) ConnectToServer(endpointConfig *Config.TcpEndpoint, transient bool) error {
	return client.attemptServerConnection(endpointConfig, transient)
}

// Closes the connection to a Systemge Server
func (client *SystemgeClient) DisconnectFromServer(address string) error {
	client.serverConnectionMutex.Lock()
	defer client.serverConnectionMutex.Unlock()
	if serverConnectionAttempt := client.serverConnectionAttempts[address]; serverConnectionAttempt != nil {
		serverConnectionAttempt.isAborted = true
	}
	if serverConnection := client.serverConnections[address]; serverConnection != nil {
		serverConnection.netConn.Close()
		serverConnection.isTransient = true
	}
	return nil
}
