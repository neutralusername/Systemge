package Client

import "Systemge/Error"

func (client *Client) addBrokerConnection(brokerConnection *brokerConnection) error {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	if client.activeBrokerConnections[brokerConnection.resolution.GetAddress()] != nil {
		return Error.New("Server connection already exists", nil)
	}
	client.activeBrokerConnections[brokerConnection.resolution.GetAddress()] = brokerConnection
	go client.handleBrokerMessages(brokerConnection)
	go client.heartbeatLoop(brokerConnection)
	return nil
}

func (client *Client) getBrokerConnection(brokerAddress string) *brokerConnection {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	return client.activeBrokerConnections[brokerAddress]
}

// Closes and removes a broker connection from the client
func (client *Client) RemoveBrokerConnection(brokerAddress string) error {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	brokerConnection := client.activeBrokerConnections[brokerAddress]
	if brokerConnection == nil {
		return Error.New("Server connection does not exist", nil)
	}
	err := brokerConnection.close()
	if err != nil {
		return Error.New("Error closing server connection", err)
	}
	delete(client.activeBrokerConnections, brokerAddress)
	return nil
}

func (client *Client) removeAllBrokerConnections() {
	client.clientMutex.Lock()
	defer client.clientMutex.Unlock()
	for address, brokerConnection := range client.activeBrokerConnections {
		brokerConnection.close()
		delete(client.activeBrokerConnections, address)
		for topic := range brokerConnection.topics {
			delete(client.topicResolutions, topic)
		}
	}
}
