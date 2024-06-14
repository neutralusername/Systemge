package MessageBrokerClient

import (
	"Systemge/Error"
	"Systemge/ResolverServer"
)

func (client *Client) addServerConnection(serverConnection *serverConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.serverConnections[serverConnection.broker.Name] != nil {
		return Error.New("Server connection already exists", nil)
	}
	client.serverConnections[serverConnection.broker.Name] = serverConnection
	go client.handleServerMessages(serverConnection)
	go client.heartbeatLoop(serverConnection)
	return nil
}

func (client *Client) getServerConnection(broker *ResolverServer.Broker) *serverConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.serverConnections[broker.Name]
}
