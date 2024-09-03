package BrokerClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/TcpConnection"
)

type getBrokerConnectionAttempt struct {
	ongoing    chan bool
	connection *connection
	err        error
}

func (messageBrokerClient *Client) getBrokerConnection(endpoint *Config.TcpClient, stopChannel chan bool) (*connection, error) {
	messageBrokerClient.mutex.Lock()
	if conn := messageBrokerClient.brokerConnections[getEndpointString(endpoint)]; conn != nil {
		messageBrokerClient.mutex.Unlock()
		return conn, nil
	}
	if ongoingGetBrokerAttempt := messageBrokerClient.ongoingGetBrokerConnections[getEndpointString(endpoint)]; ongoingGetBrokerAttempt != nil {
		messageBrokerClient.mutex.Unlock()
		<-ongoingGetBrokerAttempt.ongoing
		return ongoingGetBrokerAttempt.connection, ongoingGetBrokerAttempt.err
	}
	getBrokerAttempt := &getBrokerConnectionAttempt{
		ongoing: make(chan bool),
	}
	messageBrokerClient.ongoingGetBrokerConnections[getEndpointString(endpoint)] = getBrokerAttempt
	messageBrokerClient.mutex.Unlock()

	systemgeConnection, err := TcpConnection.EstablishConnection(messageBrokerClient.config.ConnectionConfig, endpoint, messageBrokerClient.GetName(), messageBrokerClient.config.MaxServerNameLength)
	if err != nil {
		messageBrokerClient.mutex.Lock()
		getBrokerAttempt.err = err
		delete(messageBrokerClient.ongoingGetBrokerConnections, getEndpointString(endpoint))
		close(getBrokerAttempt.ongoing)
		messageBrokerClient.mutex.Unlock()

		return nil, err
	}
	conn := &connection{
		connection:             systemgeConnection,
		endpoint:               endpoint,
		responsibleAsyncTopics: make(map[string]bool),
		responsibleSyncTopics:  make(map[string]bool),
	}

	messageBrokerClient.mutex.Lock()
	messageBrokerClient.brokerConnections[getEndpointString(endpoint)] = conn
	getBrokerAttempt.connection = conn
	delete(messageBrokerClient.ongoingGetBrokerConnections, getEndpointString(endpoint))
	close(getBrokerAttempt.ongoing)
	go messageBrokerClient.handleConnectionLifetime(conn, stopChannel)
	messageBrokerClient.mutex.Unlock()

	return conn, nil
}
