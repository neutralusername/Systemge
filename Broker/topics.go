package Broker

func (broker *Broker) addAsyncTopics(topics ...string) {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	for _, topic := range topics {
		broker.asyncTopics[topic] = true
		broker.nodeSubscriptions[topic] = map[string]*nodeConnection{}
	}
}

func (broker *Broker) addSyncTopics(topics ...string) {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	for _, topic := range topics {
		broker.syncTopics[topic] = true
		broker.nodeSubscriptions[topic] = map[string]*nodeConnection{}
	}
}

func (broker *Broker) removeAsyncTopics(topics ...string) {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	for _, topic := range topics {
		delete(broker.asyncTopics, topic)
		for _, nodeConnection := range broker.nodeSubscriptions[topic] {
			delete(nodeConnection.subscribedTopics, topic)
			delete(broker.nodeSubscriptions[topic], nodeConnection.name)
		}
		delete(broker.nodeSubscriptions, topic)
	}
}

func (broker *Broker) removeSyncTopics(topics ...string) {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	for _, topic := range topics {
		delete(broker.syncTopics, topic)
		for _, nodeConnection := range broker.nodeSubscriptions[topic] {
			delete(nodeConnection.subscribedTopics, topic)
			delete(broker.nodeSubscriptions[topic], nodeConnection.name)
		}
		delete(broker.nodeSubscriptions, topic)
	}
}
