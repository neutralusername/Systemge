package Client

import "Systemge/Utilities"

func (client *Client) addBrokerConnection(brokerConnection *brokerConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.activeBrokerConnections[brokerConnection.resolution.Address] != nil {
		return Utilities.NewError("Server connection already exists", nil)
	}
	client.activeBrokerConnections[brokerConnection.resolution.Address] = brokerConnection
	go client.handleBrokerMessages(brokerConnection)
	go client.heartbeatLoop(brokerConnection)
	return nil
}

func (client *Client) getBrokerConnection(brokerAddress string) *brokerConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.activeBrokerConnections[brokerAddress]
}
