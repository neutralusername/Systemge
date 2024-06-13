package MessageBrokerServer

// adds topics the server will accept messages and subscriptions for
func (server *Server) AddTopics(topics ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.subscriptions[topic] = make(map[string]*Client)
	}
}

// removes topics the server will accept messages and subscriptions for
func (server *Server) RemoveTopics(topics ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		for _, client := range server.subscriptions[topic] {
			client.disconnect()
		}
		delete(server.subscriptions, topic)
	}
}
