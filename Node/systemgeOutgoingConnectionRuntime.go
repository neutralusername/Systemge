package Node

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
)

// Returns a slice of addresses of nodes that this node is connected to.
func (node *Node) GetOutgoingConnectionsList() []string {
	if systemge := node.systemge; systemge != nil {
		systemge.outgoingConnectionMutex.Lock()
		defer systemge.outgoingConnectionMutex.Unlock()
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
	if systemge := node.systemge; systemge != nil {
		systemge.outgoingConnectionMutex.Lock()
		defer systemge.outgoingConnectionMutex.Unlock()
		attempts := make([]string, len(systemge.currentlyInOutgoingConnectionLoop))
		i := 0
		for address := range systemge.currentlyInOutgoingConnectionLoop {
			attempts[i] = address
			i++
		}
		return attempts
	}
	return nil
}

// Adds another node as an outgoing connection.
// This connection is used to send async and sync requests and receive sync responses for their corresponding requests.
func (node *Node) ConnectToNode(endpointConfig *Config.TcpEndpoint) error {
	if systemge := node.systemge; systemge != nil {
		systemge.outgoingConnectionMutex.Lock()
		if systemge.outgoingConnections[endpointConfig.Address] != nil {
			systemge.outgoingConnectionMutex.Unlock()
			return Error.New("Connection to endpoint \""+endpointConfig.Address+"\" already exists", nil)
		}
		if systemge.currentlyInOutgoingConnectionLoop[endpointConfig.Address] != nil {
			systemge.outgoingConnectionMutex.Unlock()
			return Error.New("Connection to endpoint \""+endpointConfig.Address+"\" is already being established", nil)
		}
		b := true
		systemge.currentlyInOutgoingConnectionLoop[endpointConfig.Address] = &b
		systemge.outgoingConnectionMutex.Unlock()
		return node.outgoingConnectionLoop(endpointConfig)
	}
	return Error.New("Systemge is nil", nil)
}

// Removes a node from the outgoing connections and aborts ongoing connection attempts.
func (node *Node) DisconnectFromNode(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.outgoingConnectionMutex.Lock()
		defer systemge.outgoingConnectionMutex.Unlock()
		if systemge.currentlyInOutgoingConnectionLoop[address] != nil {
			*systemge.currentlyInOutgoingConnectionLoop[address] = false
		}
		if outgoingConnection := systemge.outgoingConnections[address]; outgoingConnection != nil {
			outgoingConnection.netConn.Close()
			outgoingConnection.transient = true
		}
		return nil
	}
	return Error.New("Systemge is nil", nil)
}
