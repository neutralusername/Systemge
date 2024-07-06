package Node

func (node *Node) handleBrokerDisconnect(brokerConnection *brokerConnection) {
	removedSubscribedTopics := node.cleanUpDisconnectedBrokerConnection(brokerConnection)
	for _, topic := range removedSubscribedTopics {
		go node.subscribeLoop(topic)
	}
}

func (node *Node) cleanUpDisconnectedBrokerConnection(brokerConnection *brokerConnection) []string {
	node.mutex.Lock()
	brokerConnection.mutex.Lock()
	delete(node.brokerConnections, brokerConnection.endpoint.GetAddress())
	removedSubscribedTopics := make([]string, 0)
	for topic := range brokerConnection.topics {
		delete(node.topicBrokerConnections, topic)
		if node.application.GetAsyncMessageHandlers()[topic] != nil || node.application.GetSyncMessageHandlers()[topic] != nil {
			removedSubscribedTopics = append(removedSubscribedTopics, topic)
		}
	}
	brokerConnection.topics = make(map[string]bool)
	brokerConnection.mutex.Unlock()
	node.mutex.Unlock()
	return removedSubscribedTopics
}
