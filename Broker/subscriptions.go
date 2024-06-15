package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
)

func (server *Server) addSubscription(clientConnection *clientConnection, topic string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if !server.asyncTopics[topic] && !server.syncTopics[topic] {
		return Error.New("Topic \""+topic+"\" does not exist on server \""+server.name+"\"", nil)
	}
	if clientConnection.subscribedTopics[topic] {
		return Error.New("Client \""+clientConnection.name+"\" is already subscribed to topic \""+topic+"\"", nil)
	}
	if server.clientSubscriptions[topic] == nil {
		return Error.New("Topic \""+topic+"\" does not exist", nil)
	}
	server.clientSubscriptions[topic][clientConnection.name] = clientConnection
	clientConnection.subscribedTopics[topic] = true
	return nil
}

func (server *Server) removeSubscription(clientConnection *clientConnection, topic string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.clientSubscriptions[topic] == nil {
		return Error.New("Topic \""+topic+"\" does not exist", nil)
	}
	if !clientConnection.subscribedTopics[topic] {
		return Error.New("Client \""+clientConnection.name+"\" is not subscribed to topic \""+topic+"\"", nil)
	}
	delete(server.clientSubscriptions[topic], clientConnection.name)
	delete(clientConnection.subscribedTopics, topic)
	if len(server.clientSubscriptions[topic]) == 0 {
		delete(server.clientSubscriptions, topic)
	}
	return nil
}

func (server *Server) getSubscribers(topic string) []*clientConnection {
	subscribers := []*clientConnection{}
	for _, clientConnection := range server.clientSubscriptions[topic] {
		subscribers = append(subscribers, clientConnection)
	}
	return subscribers
}

func (server *Server) propagateMessage(clients []*clientConnection, message *Message.Message) {
	for _, clientConnection := range clients {
		if clientConnection.deliverImmediately {
			err := clientConnection.send(message)
			if err != nil {
				server.logger.Log(Error.New("Failed to send message to client \""+clientConnection.name+"\"", err).Error())
			}
		} else {
			clientConnection.queueMessage(message)
		}
	}
}
