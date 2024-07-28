package Node

func (broker *brokerComponent) addAsyncTopics(topics ...string) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()
	for _, topic := range topics {
		broker.asyncTopics[topic] = true
		broker.nodeSubscriptions[topic] = map[string]*nodeConnection{}
	}
}

func (broker *brokerComponent) addSyncTopics(topics ...string) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()
	for _, topic := range topics {
		broker.syncTopics[topic] = true
		broker.nodeSubscriptions[topic] = map[string]*nodeConnection{}
	}
}

func (broker *brokerComponent) removeAsyncTopics(topics ...string) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()
	for _, topic := range topics {
		delete(broker.asyncTopics, topic)
		for _, nodeConnection := range broker.nodeSubscriptions[topic] {
			delete(nodeConnection.subscribedTopics, topic)
			delete(broker.nodeSubscriptions[topic], nodeConnection.name)
		}
		delete(broker.nodeSubscriptions, topic)
	}
}

func (broker *brokerComponent) removeSyncTopics(topics ...string) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()
	for _, topic := range topics {
		delete(broker.syncTopics, topic)
		for _, nodeConnection := range broker.nodeSubscriptions[topic] {
			delete(nodeConnection.subscribedTopics, topic)
			delete(broker.nodeSubscriptions[topic], nodeConnection.name)
		}
		delete(broker.nodeSubscriptions, topic)
	}
}
