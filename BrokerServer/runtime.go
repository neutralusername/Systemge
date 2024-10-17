package BrokerServer

import (
	"github.com/neutralusername/systemge/SystemgeConnection"
)

func (server *Server) AddAsyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.messageHandler.AddAsyncMessageHandler(topic, server.handleAsyncPropagate)
		if server.asyncTopicSubscriptions[topic] == nil {
			server.asyncTopicSubscriptions[topic] = make(map[SystemgeConnection.SystemgeConnection]bool)
		}
	}
}

func (server *Server) AddSyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.messageHandler.AddSyncMessageHandler(topic, server.handleSyncPropagate)
		if server.syncTopicSubscriptions[topic] == nil {
			server.syncTopicSubscriptions[topic] = make(map[SystemgeConnection.SystemgeConnection]bool)
		}
	}
}

func (server *Server) RemoveAsyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicSubscriptions[topic] != nil {
			server.messageHandler.RemoveAsyncMessageHandler(topic)
			delete(server.asyncTopicSubscriptions, topic)
		}
	}
}

func (server *Server) RemoveSyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicSubscriptions[topic] != nil {
			server.messageHandler.RemoveSyncMessageHandler(topic)
			delete(server.syncTopicSubscriptions, topic)
		}
	}
}
