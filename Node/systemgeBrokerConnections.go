package Node

import (
	"Systemge/Error"
)

func (node *Node) addBrokerConnection(brokerConnection *brokerConnection) error {
	node.systemgeMutex.Lock()
	defer node.systemgeMutex.Unlock()
	if node.systemgeBrokerConnections[brokerConnection.endpoint.Address] != nil {
		return Error.New("broker connection already exists", nil)
	}
	node.systemgeBrokerConnections[brokerConnection.endpoint.Address] = brokerConnection
	go node.handleSystemgeMessages(brokerConnection)
	return nil
}

func (node *Node) getBrokerConnection(brokerAddress string) *brokerConnection {
	node.systemgeMutex.Lock()
	defer node.systemgeMutex.Unlock()
	return node.systemgeBrokerConnections[brokerAddress]
}

func (node *Node) removeAllBrokerConnections() {
	node.systemgeMutex.Lock()
	defer node.systemgeMutex.Unlock()
	for address, brokerConnection := range node.systemgeBrokerConnections {
		brokerConnection.close()
		brokerConnection.mutex.Lock()
		delete(node.systemgeBrokerConnections, address)
		for topic := range brokerConnection.topicResolutions {
			delete(node.systemgeTopicResolutions, topic)
		}
		for topic := range brokerConnection.subscribedTopics {
			delete(brokerConnection.subscribedTopics, topic)
		}
		brokerConnection.mutex.Unlock()
	}
}
