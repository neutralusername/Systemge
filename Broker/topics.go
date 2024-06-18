package Broker

import (
	"Systemge/Message"
	"Systemge/Utilities"
)

func (server *Server) validateMessageTopic(message *Message.Message) error {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if message.GetSyncRequestToken() != "" {
		if !server.syncTopics[message.GetTopic()] {
			return Utilities.NewError("Sync Topic \""+message.GetTopic()+"\" does not exist on server \""+server.name+"\"", nil)
		}
	} else {
		if !server.asyncTopics[message.GetTopic()] {
			return Utilities.NewError("Async Topic \""+message.GetTopic()+"\" does not exist on server \""+server.name+"\"", nil)
		}
	}
	return nil
}

// adds topics the server will accept async messages and subscriptions for
func (server *Server) AddAsyncTopics(topics ...string) {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, topic := range topics {
		server.asyncTopics[topic] = true
		server.clientSubscriptions[topic] = map[string]*clientConnection{}
	}
}

// adds topics the server will accept sync messages and subscriptions for
func (server *Server) AddSyncTopics(topics ...string) {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, topic := range topics {
		server.syncTopics[topic] = true
		server.clientSubscriptions[topic] = map[string]*clientConnection{}
	}
}

// removes topics the server will accept async messages and subscriptions for
func (server *Server) RemoveAsyncTopics(topics ...string) {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, topic := range topics {
		delete(server.asyncTopics, topic)
		for _, client := range server.clientSubscriptions[topic] {
			delete(client.subscribedTopics, topic)
			delete(server.clientSubscriptions[topic], client.name)
		}
		delete(server.clientSubscriptions, topic)
	}
}

// removes topics the server will accept sync messages and subscriptions for
func (server *Server) RemoveSyncTopics(topics ...string) {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	for _, topic := range topics {
		delete(server.syncTopics, topic)
		for _, client := range server.clientSubscriptions[topic] {
			delete(client.subscribedTopics, topic)
			delete(server.clientSubscriptions[topic], client.name)
		}
		delete(server.clientSubscriptions, topic)
	}
}
