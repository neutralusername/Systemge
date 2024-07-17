package Node

func (node *Node) handleBrokerDisconnect(brokerConnection *brokerConnection) {
	brokerConnection.close()
	removedSubscribedTopics := node.cleanUpDisconnectedBrokerConnection(brokerConnection)
	for _, topic := range removedSubscribedTopics {
		go node.subscribeLoop(topic)
	}
}

func (node *Node) cleanUpDisconnectedBrokerConnection(brokerConnection *brokerConnection) []string {
	node.systemgeMutex.Lock()
	defer node.systemgeMutex.Unlock()
	brokerConnection.mutex.Lock()
	delete(node.systemgeBrokerConnections, brokerConnection.endpoint.GetAddress())
	removedSubscribedTopics := make([]string, 0)
	for topic := range brokerConnection.topicResolutions {
		delete(node.systemgeTopicResolutions, topic)
	}
	for topic := range brokerConnection.subscribedTopics {
		removedSubscribedTopics = append(removedSubscribedTopics, topic)
		delete(brokerConnection.subscribedTopics, topic)
	}
	brokerConnection.topicResolutions = make(map[string]bool)
	brokerConnection.mutex.Unlock()
	return removedSubscribedTopics
}
