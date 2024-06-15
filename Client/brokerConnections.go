package Client

import (
	"Systemge/Error"
)

func (client *Client) addBrokerConnection(brokerConnection *brokerConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.activeBrokerConnections[brokerConnection.resolution.Address] != nil {
		return Error.New("Server connection already exists", nil)
	}
	client.activeBrokerConnections[brokerConnection.resolution.Address] = brokerConnection
	go client.handleServerMessages(brokerConnection)
	go client.heartbeatLoop(brokerConnection)
	return nil
}

func (client *Client) getBrokerConnection(brokerAddress string) *brokerConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.activeBrokerConnections[brokerAddress]
}
