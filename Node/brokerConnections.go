package Node

import "Systemge/Error"

func (node *Node) addBrokerConnection(brokerConnection *brokerConnection) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if node.brokerConnections[brokerConnection.resolution.GetAddress()] != nil {
		return Error.New("broker connection already exists", nil)
	}
	node.brokerConnections[brokerConnection.resolution.GetAddress()] = brokerConnection
	go node.handleBrokerMessages(brokerConnection)
	go node.heartbeatLoop(brokerConnection)
	return nil
}

func (node *Node) getBrokerConnection(brokerAddress string) *brokerConnection {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	return node.brokerConnections[brokerAddress]
}

// Closes and removes a broker connection from the node
func (node *Node) RemoveBrokerConnection(brokerAddress string) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	brokerConnection := node.brokerConnections[brokerAddress]
	if brokerConnection == nil {
		return Error.New("broker connection does not exist", nil)
	}
	err := brokerConnection.close()
	if err != nil {
		return Error.New("Error closing broker connection", err)
	}
	delete(node.brokerConnections, brokerAddress)
	for topic := range brokerConnection.topics {
		delete(node.topicResolutions, topic)
	}
	return nil
}

func (node *Node) removeAllBrokerConnections() {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	for address, brokerConnection := range node.brokerConnections {
		brokerConnection.close()
		delete(node.brokerConnections, address)
		for topic := range brokerConnection.topics {
			delete(node.topicResolutions, topic)
		}
	}
}
