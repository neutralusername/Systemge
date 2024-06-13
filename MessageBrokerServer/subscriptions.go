package MessageBrokerServer

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (server *Server) addSubscription(client *Client, topic string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.subscriptions[topic] == nil {
		return Error.New("Topic \""+topic+"\" does not exist", nil)
	}
	if client.subscribedTopics[topic] {
		return Error.New("Client \""+client.name+"\" is already subscribed to topic \""+topic+"\"", nil)
	}
	server.subscriptions[topic][client.name] = client
	client.subscribedTopics[topic] = true
	return nil
}

func (server *Server) removeSubscription(client *Client, topic string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.subscriptions[topic] == nil {
		return Error.New("Topic \""+topic+"\" does not exist", nil)
	}
	if !client.subscribedTopics[topic] {
		return Error.New("Client \""+client.name+"\" is not subscribed to topic \""+topic+"\"", nil)
	}
	delete(server.subscriptions[topic], client.name)
	delete(client.subscribedTopics, topic)
	return nil
}

func (server *Server) getSubscribers(topic string) []*Client {
	subscribers := []*Client{}
	for _, client := range server.subscriptions[topic] {
		subscribers = append(subscribers, client)
	}
	return subscribers
}

func (server *Server) propagateMessage(clients []*Client, message *Message.Message) {
	for _, client := range clients {
		if client.deliverImmediately {
			err := client.send(message)
			if err != nil {
				server.logger.Log(Error.New("Failed to send message to client \""+client.name+"\"", err).Error())
			}
		} else {
			client.queueMessage(message)
		}
	}
}
