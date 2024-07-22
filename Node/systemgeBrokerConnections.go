package Node

import (
	"Systemge/Error"
)

func (systemge *systemgeComponent) addBrokerConnection(brokerConnection *brokerConnection) error {
	systemge.mutex.Lock()
	defer systemge.mutex.Unlock()
	if systemge.brokerConnections[brokerConnection.endpoint.Address] != nil {
		return Error.New("broker connection already exists", nil)
	}
	systemge.brokerConnections[brokerConnection.endpoint.Address] = brokerConnection
	return nil
}

func (systemge *systemgeComponent) getBrokerConnection(brokerAddress string) *brokerConnection {
	systemge.mutex.Lock()
	defer systemge.mutex.Unlock()
	return systemge.brokerConnections[brokerAddress]
}

func (systemge *systemgeComponent) removeAllBrokerConnections() {
	systemge.mutex.Lock()
	defer systemge.mutex.Unlock()
	for address, brokerConnection := range systemge.brokerConnections {
		brokerConnection.close()
		brokerConnection.mutex.Lock()
		delete(systemge.brokerConnections, address)
		for topic := range brokerConnection.topicResolutions {
			delete(systemge.topicResolutions, topic)
		}
		for topic := range brokerConnection.subscribedTopics {
			delete(brokerConnection.subscribedTopics, topic)
		}
		brokerConnection.mutex.Unlock()
	}
}
