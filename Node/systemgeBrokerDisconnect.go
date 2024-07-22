package Node

func (systemge *systemgeComponent) cleanUpDisconnectedBrokerConnection(brokerConnection *brokerConnection) []string {
	systemge.mutex.Lock()
	brokerConnection.mutex.Lock()
	defer func() {
		brokerConnection.mutex.Unlock()
		systemge.mutex.Unlock()
	}()
	defer delete(systemge.brokerConnections, brokerConnection.endpoint.Address)
	removedSubscribedTopics := make([]string, 0)
	for topic := range brokerConnection.topicResolutions {
		delete(systemge.topicResolutions, topic)
	}
	for topic := range brokerConnection.subscribedTopics {
		removedSubscribedTopics = append(removedSubscribedTopics, topic)
		delete(brokerConnection.subscribedTopics, topic)
	}
	brokerConnection.topicResolutions = make(map[string]bool)

	return removedSubscribedTopics
}
