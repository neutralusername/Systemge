package Client

import (
	"Systemge/Error"
)

func (client *Client) addServerConnection(serverConnection *brokerConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.activeBrokerConnections[serverConnection.resolution.Address] != nil {
		return Error.New("Server connection already exists", nil)
	}
	client.activeBrokerConnections[serverConnection.resolution.Address] = serverConnection
	go client.handleServerMessages(serverConnection)
	go client.heartbeatLoop(serverConnection)
	return nil
}

func (client *Client) getServerConnection(brokerAddress string) *brokerConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.activeBrokerConnections[brokerAddress]
}
