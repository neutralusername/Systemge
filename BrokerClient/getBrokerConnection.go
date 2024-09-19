package BrokerClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
)

type getBrokerConnectionAttempt struct {
	ongoing    chan bool
	connection *connection
	err        error
}

func (messageBrokerClient *Client) getBrokerConnection(tcpClientConfig *Config.TcpClient, stopChannel chan bool) (*connection, error) {
	messageBrokerClient.mutex.Lock()
	if conn := messageBrokerClient.brokerConnections[getTcpClientConfigString(tcpClientConfig)]; conn != nil {
		messageBrokerClient.mutex.Unlock()
		return conn, nil
	}
	if ongoingGetBrokerAttempt := messageBrokerClient.ongoingGetBrokerConnections[getTcpClientConfigString(tcpClientConfig)]; ongoingGetBrokerAttempt != nil {
		messageBrokerClient.mutex.Unlock()
		<-ongoingGetBrokerAttempt.ongoing
		return ongoingGetBrokerAttempt.connection, ongoingGetBrokerAttempt.err
	}
	getBrokerAttempt := &getBrokerConnectionAttempt{
		ongoing: make(chan bool),
	}
	messageBrokerClient.ongoingGetBrokerConnections[getTcpClientConfigString(tcpClientConfig)] = getBrokerAttempt
	messageBrokerClient.mutex.Unlock()

	connectionAttempt := TcpSystemgeConnection.EstablishConnectionAttempts(messageBrokerClient.name, &Config.SystemgeConnectionAttempt{
		MaxServerNameLength:         messageBrokerClient.config.MaxServerNameLength,
		TcpClientConfig:             tcpClientConfig,
		TcpSystemgeConnectionConfig: messageBrokerClient.config.ServerTcpSystemgeConnectionConfig,
		RetryIntervalMs:             messageBrokerClient.config.ResolutionAttemptRetryIntervalMs,
		MaxConnectionAttempts:       messageBrokerClient.config.ResolutionMaxAttempts,
	})
	select {
	case <-connectionAttempt.GetOngoingChannel():
	case <-stopChannel:
		connectionAttempt.AbortAttempts()
	}
	systemgeConnection, err := connectionAttempt.GetResultBlocking()
	if err != nil {
		messageBrokerClient.mutex.Lock()
		getBrokerAttempt.err = err
		delete(messageBrokerClient.ongoingGetBrokerConnections, getTcpClientConfigString(tcpClientConfig))
		close(getBrokerAttempt.ongoing)
		messageBrokerClient.mutex.Unlock()

		return nil, err
	}
	err = systemgeConnection.StartMessageHandlingLoop_Concurrently(messageBrokerClient.messageHandler)
	if err != nil {
		messageBrokerClient.mutex.Lock()
		getBrokerAttempt.err = err
		delete(messageBrokerClient.ongoingGetBrokerConnections, getTcpClientConfigString(tcpClientConfig))
		close(getBrokerAttempt.ongoing)
		messageBrokerClient.mutex.Unlock()

		return nil, err
	}
	conn := &connection{
		connection:             systemgeConnection,
		tcpClientConfig:        tcpClientConfig,
		responsibleAsyncTopics: make(map[string]bool),
		responsibleSyncTopics:  make(map[string]bool),
	}

	messageBrokerClient.mutex.Lock()
	messageBrokerClient.brokerConnections[getTcpClientConfigString(tcpClientConfig)] = conn
	getBrokerAttempt.connection = conn
	delete(messageBrokerClient.ongoingGetBrokerConnections, getTcpClientConfigString(tcpClientConfig))
	close(getBrokerAttempt.ongoing)
	go messageBrokerClient.handleConnectionLifetime(conn, stopChannel)
	messageBrokerClient.mutex.Unlock()

	return conn, nil
}
