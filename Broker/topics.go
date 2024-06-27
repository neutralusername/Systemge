package Broker

// adds topics the broker will accept async messages and subscriptions for
func (broker *Broker) AddAsyncTopics(topics ...string) {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	for _, topic := range topics {
		broker.asyncTopics[topic] = true
		broker.nodeSubscriptions[topic] = map[string]*nodeConnection{}
	}
}

// adds topics the broker will accept sync messages and subscriptions for
func (broker *Broker) AddSyncTopics(topics ...string) {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	for _, topic := range topics {
		broker.syncTopics[topic] = true
		broker.nodeSubscriptions[topic] = map[string]*nodeConnection{}
	}
}

// removes topics the broker will accept async messages and subscriptions for
func (broker *Broker) RemoveAsyncTopics(topics ...string) {
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

// removes topics the broker will accept sync messages and subscriptions for
func (broker *Broker) RemoveSyncTopics(topics ...string) {
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
