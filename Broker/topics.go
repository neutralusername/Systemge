package Broker

// adds topics the server will accept async messages and subscriptions for
func (server *Server) AddAsyncTopics(topics ...string) {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, topic := range topics {
		server.asyncTopics[topic] = true
		server.nodeSubscriptions[topic] = map[string]*nodeConnection{}
	}
}

// adds topics the server will accept sync messages and subscriptions for
func (server *Server) AddSyncTopics(topics ...string) {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, topic := range topics {
		server.syncTopics[topic] = true
		server.nodeSubscriptions[topic] = map[string]*nodeConnection{}
	}
}

// removes topics the server will accept async messages and subscriptions for
func (server *Server) RemoveAsyncTopics(topics ...string) {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, topic := range topics {
		delete(server.asyncTopics, topic)
		for _, nodeConnection := range server.nodeSubscriptions[topic] {
			delete(nodeConnection.subscribedTopics, topic)
			delete(server.nodeSubscriptions[topic], nodeConnection.name)
		}
		delete(server.nodeSubscriptions, topic)
	}
}

// removes topics the server will accept sync messages and subscriptions for
func (server *Server) RemoveSyncTopics(topics ...string) {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, topic := range topics {
		delete(server.syncTopics, topic)
		for _, nodeConnection := range server.nodeSubscriptions[topic] {
			delete(nodeConnection.subscribedTopics, topic)
			delete(server.nodeSubscriptions[topic], nodeConnection.name)
		}
		delete(server.nodeSubscriptions, topic)
	}
}
