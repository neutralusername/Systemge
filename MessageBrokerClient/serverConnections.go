package MessageBrokerClient

import (
	"Systemge/Error"
	"Systemge/ResolverServer"
)

func (client *Client) addServerConnection(serverConnection *serverConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.serverConnections[serverConnection.resolution.Name] != nil {
		return Error.New("Server connection already exists", nil)
	}
	client.serverConnections[serverConnection.resolution.Name] = serverConnection
	go client.handleServerMessages(serverConnection)
	go client.heartbeatLoop(serverConnection)
	return nil
}

func (client *Client) getServerConnection(resolution *ResolverServer.Resolution) *serverConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.serverConnections[resolution.Name]
}
