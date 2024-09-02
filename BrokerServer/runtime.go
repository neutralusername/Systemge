package BrokerServer

import "github.com/neutralusername/Systemge/SystemgeConnection"

func (server *MessageBrokerServer) AddAsyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.messageHandler.AddAsyncMessageHandler(topic, server.handleAsyncPropagate)
		if server.asyncTopicSubscriptions[topic] == nil {
			server.asyncTopicSubscriptions[topic] = make(map[*SystemgeConnection.SystemgeConnection]bool)
		}
	}
}

func (server *MessageBrokerServer) AddSyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.messageHandler.AddSyncMessageHandler(topic, server.handleSyncPropagate)
		if server.syncTopicSubscriptions[topic] == nil {
			server.syncTopicSubscriptions[topic] = make(map[*SystemgeConnection.SystemgeConnection]bool)
		}
	}
}

func (server *MessageBrokerServer) RemoveAsyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicSubscriptions[topic] != nil {
			server.messageHandler.RemoveAsyncMessageHandler(topic)
			delete(server.asyncTopicSubscriptions, topic)
		}
	}
}

func (server *MessageBrokerServer) RemoveSyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicSubscriptions[topic] != nil {
			server.messageHandler.RemoveSyncMessageHandler(topic)
			delete(server.syncTopicSubscriptions, topic)
		}
	}
}
