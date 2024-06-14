package MessageBrokerServer

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (server *Server) validateMessageTopic(message *Message.Message) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.asyncTopics[message.GetTopic()] {
		if message.GetSyncRequestToken() != "" {
			return Error.New("broker does not accept sync messages for topic \""+message.GetTopic()+"\"", nil)
		}
	} else if server.syncTopics[message.GetTopic()] {
		if message.GetSyncRequestToken() == "" {
			return Error.New("broker only accepts sync messages for topic \""+message.GetTopic()+"\"", nil)
		}
	} else {
		return Error.New("Topic \""+message.GetTopic()+"\" does not exist on server \""+server.name+"\"", nil)
	}
	return nil
}

// adds topics the server will accept async messages and subscriptions for
func (server *Server) AddAsyncTopics(topics ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.asyncTopics[topic] = true
	}
}

// adds topics the server will accept sync messages and subscriptions for
func (server *Server) AddSyncTopics(topics ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.syncTopics[topic] = true
	}
}

// removes topics the server will accept async messages and subscriptions for
func (server *Server) RemoveAsyncTopics(topics ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		delete(server.asyncTopics, topic)
		if server.subscriptions[topic] != nil {
			for _, client := range server.subscriptions[topic] {
				delete(client.subscribedTopics, topic)
				delete(server.subscriptions[topic], client.name)
			}
		}
	}
}

// removes topics the server will accept sync messages and subscriptions for
func (server *Server) RemoveSyncTopics(topics ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		delete(server.syncTopics, topic)
		if server.subscriptions[topic] != nil {
			for _, client := range server.subscriptions[topic] {
				delete(client.subscribedTopics, topic)
				delete(server.subscriptions[topic], client.name)
			}
		}
	}
}
