package Client

import "Systemge/Utilities"

func (client *Client) addBrokerConnection(brokerConnection *brokerConnection) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	if client.activeBrokerConnections[brokerConnection.resolution.GetAddress()] != nil {
		return Utilities.NewError("Server connection already exists", nil)
	}
	client.activeBrokerConnections[brokerConnection.resolution.GetAddress()] = brokerConnection
	go client.handleBrokerMessages(brokerConnection)
	go client.heartbeatLoop(brokerConnection)
	return nil
}

func (client *Client) getBrokerConnection(brokerAddress string) *brokerConnection {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	return client.activeBrokerConnections[brokerAddress]
}

// Closes and removes a broker connection from the client
func (client *Client) RemoveBrokerConnection(brokerAddress string) error {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	brokerConnection := client.activeBrokerConnections[brokerAddress]
	if brokerConnection == nil {
		return Utilities.NewError("Server connection does not exist", nil)
	}
	err := brokerConnection.close()
	if err != nil {
		return Utilities.NewError("Error closing server connection", err)
	}
	delete(client.activeBrokerConnections, brokerAddress)
	return nil
}
