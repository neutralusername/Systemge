package Node

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
)

// Returns a slice of addresses of nodes that this node is connected to.
func (node *Node) GetOutgoingConnectionsList() []string {
	if systemge := node.systemgeClient; systemge != nil {
		systemge.outgoingConnectionMutex.RLock()
		defer systemge.outgoingConnectionMutex.RUnlock()
		connections := make([]string, len(systemge.outgoingConnections))
		i := 0
		for address := range systemge.outgoingConnections {
			connections[i] = address
			i++
		}
		return connections
	}
	return nil
}

// Returns a slice of addresses of nodes that this node is currently trying to connect to.
func (node *Node) GetOutgoingConnectionAttemptsList() []string {
	if systemge := node.systemgeClient; systemge != nil {
		systemge.outgoingConnectionMutex.RLock()
		defer systemge.outgoingConnectionMutex.RUnlock()
		attempts := make([]string, len(systemge.outgoingConnectionAttempts))
		i := 0
		for address := range systemge.outgoingConnectionAttempts {
			attempts[i] = address
			i++
		}
		return attempts
	}
	return nil
}

// Adds another node as an outgoing connection.
// This connection is used to send async and sync requests and receive sync responses for their corresponding requests.
func (node *Node) ConnectToNode(endpointConfig *Config.TcpEndpoint, transient bool) error {
	if systemge := node.systemgeClient; systemge != nil {
		return systemge.attemptOutgoingConnection(endpointConfig, transient)
	}
	return Error.New("Systemge is nil", nil)
}

// Removes a node from the outgoing connections and aborts ongoing connection attempts.
func (node *Node) DisconnectFromNode(address string) error {
	if systemge := node.systemgeClient; systemge != nil {
		systemge.outgoingConnectionMutex.Lock()
		defer systemge.outgoingConnectionMutex.Unlock()
		if outgoingConnectionAttempt := systemge.outgoingConnectionAttempts[address]; outgoingConnectionAttempt != nil {
			outgoingConnectionAttempt.isAborted = true
		}
		if outgoingConnection := systemge.outgoingConnections[address]; outgoingConnection != nil {
			outgoingConnection.netConn.Close()
			outgoingConnection.isTransient = true
		}
		return nil
	}
	return Error.New("Systemge is nil", nil)
}
