package Node

import (
	"github.com/neutralusername/Systemge/Config"
)

// Returns a slice of addresses of nodes that this node is connected to.
func (client *SystemgeClient) GetOutgoingConnectionsList() []string {
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

// Returns a slice of addresses of nodes that this node is currently trying to connect to.
func (client *SystemgeClient) GetOutgoingConnectionAttemptsList() []string {
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

// Adds another node as an outgoing connection.
// This connection is used to send async and sync requests and receive sync responses for their corresponding requests.
func (client *SystemgeClient) ConnectToNode(endpointConfig *Config.TcpEndpoint, transient bool) error {
	return client.attemptOutgoingConnection(endpointConfig, transient)
}

// Removes a node from the outgoing connections and aborts ongoing connection attempts.
func (client *SystemgeClient) DisconnectFromNode(address string) error {
	client.serverConnectionMutex.Lock()
	defer client.serverConnectionMutex.Unlock()
	if outgoingConnectionAttempt := client.serverConnectionAttempts[address]; outgoingConnectionAttempt != nil {
		outgoingConnectionAttempt.isAborted = true
	}
	if outgoingConnection := client.serverConnections[address]; outgoingConnection != nil {
		outgoingConnection.netConn.Close()
		outgoingConnection.isTransient = true
	}
	return nil
}
