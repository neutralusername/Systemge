package MessageBrokerServer

func (server *Server) GetClientCount() int {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return len(server.clients)
}

func (server *Server) GetClientNames() []string {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	names := make([]string, 0)
	for name := range server.clients {
		names = append(names, name)
	}
	return names
}

func (server *Server) GetClient(name string) *Client {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.clients[name]
}

func (client *Client) GetSubscribedTopics() []string {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	topics := make([]string, 0)
	for topic := range client.subscribedTopics {
		topics = append(topics, topic)
	}
	return topics
}

func (client *Client) IsSubscribedToTopic(topic string) bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.subscribedTopics[topic]
}

func (client *Client) GetOpenSyncRequests() []string {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	tokens := make([]string, 0)
	for token := range client.openSyncRequests {
		tokens = append(tokens, token)
	}
	return tokens
}
