package BrokerClient

/* type getBrokerConnectionAttempt struct {
	ongoing            chan bool
	systemgeConnection SystemgeConnection.SystemgeConnection
	err                error
}

func (messageBrokerClient *Client) getBrokerConnection(endpoint *Config.TcpClient, stopChannel chan bool) (SystemgeConnection.SystemgeConnection, error) {
	messageBrokerClient.mutex.Lock()
	if systemgeConnection := messageBrokerClient.brokerSystemgeClient.GetConnectionByAddress(endpoint.Address); systemgeConnection != nil {
		messageBrokerClient.mutex.Unlock()
		return systemgeConnection, nil
	}
	if ongoingGetBrokerAttempt := messageBrokerClient.ongoingGetBrokerConnections[endpoint.Address]; ongoingGetBrokerAttempt != nil {
		messageBrokerClient.mutex.Unlock()
		<-ongoingGetBrokerAttempt.ongoing
		return ongoingGetBrokerAttempt.systemgeConnection, ongoingGetBrokerAttempt.err
	}
	getBrokerAttempt := &getBrokerConnectionAttempt{
		ongoing: make(chan bool),
	}
	messageBrokerClient.ongoingGetBrokerConnections[endpoint.Address] = getBrokerAttempt
	messageBrokerClient.mutex.Unlock()

	systemgeConnection, err := TcpSystemgeConnection.EstablishConnection(messageBrokerClient.config.ServerTcpSystemgeConnectionConfig, endpoint, messageBrokerClient.GetName(), messageBrokerClient.config.MaxServerNameLength)
	if err != nil {
		messageBrokerClient.mutex.Lock()
		getBrokerAttempt.err = err
		delete(messageBrokerClient.ongoingGetBrokerConnections, getEndpointString(endpoint))
		close(getBrokerAttempt.ongoing)
		messageBrokerClient.mutex.Unlock()

		return nil, err
	}
	systemgeConnection.StartProcessingLoopConcurrently(messageBrokerClient.messageHandler)
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
*/
