package Node

import (
	"Systemge/Error"
)

func (node *Node) addBrokerConnection(brokerConnection *brokerConnection) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if node.brokerConnections[brokerConnection.endpoint.GetAddress()] != nil {
		return Error.New("broker connection already exists", nil)
	}
	node.brokerConnections[brokerConnection.endpoint.GetAddress()] = brokerConnection
	go node.handleBrokerMessages(brokerConnection)
	return nil
}

func (node *Node) getBrokerConnection(brokerAddress string) *brokerConnection {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	return node.brokerConnections[brokerAddress]
}

func (node *Node) removeAllBrokerConnections() {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	for address, brokerConnection := range node.brokerConnections {
		brokerConnection.close()
		brokerConnection.mutex.Lock()
		delete(node.brokerConnections, address)
		for topic := range brokerConnection.topicResolutions {
			delete(node.topicResolutions, topic)
		}
		for topic := range brokerConnection.subscribedTopics {
			delete(brokerConnection.subscribedTopics, topic)
		}
		brokerConnection.mutex.Unlock()
	}
}
