package Node

import (
	"github.com/neutralusername/Systemge/Error"
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
