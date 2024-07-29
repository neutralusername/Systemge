package Node

import (
	"github.com/neutralusername/Systemge/Error"
)

func (systemge *systemgeComponent) addBrokerConnection(brokerConnection *brokerConnection) error {
	systemge.brokerConnectionsMutex.Lock()
	defer systemge.brokerConnectionsMutex.Unlock()
	if systemge.brokerConnections[brokerConnection.endpoint.Address] != nil {
		return Error.New("broker connection already exists", nil)
	}
	systemge.brokerConnections[brokerConnection.endpoint.Address] = brokerConnection
	return nil
}

func (systemge *systemgeComponent) getBrokerConnection(brokerAddress string) *brokerConnection {
	systemge.brokerConnectionsMutex.Lock()
	defer systemge.brokerConnectionsMutex.Unlock()
	return systemge.brokerConnections[brokerAddress]
}
