package MessageBrokerClient

import (
	"Systemge/Error"
)

func (client *Client) addServerConnection(serverConnection *serverConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.activeServerConnections[serverConnection.resolution.Address] != nil {
		return Error.New("Server connection already exists", nil)
	}
	client.activeServerConnections[serverConnection.resolution.Address] = serverConnection
	go client.handleServerMessages(serverConnection)
	go client.heartbeatLoop(serverConnection)
	return nil
}

func (client *Client) getServerConnection(brokerAddress string) *serverConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.activeServerConnections[brokerAddress]
}
