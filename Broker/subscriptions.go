package Broker

import "Systemge/Error"

func (server *Server) addSubscription(nodeConnection *nodeConnection, topic string) error {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if !server.asyncTopics[topic] && !server.syncTopics[topic] {
		return Error.New("Topic \""+topic+"\" does not exist on server \""+server.name+"\"", nil)
	}
	if server.syncTopics[topic] && len(server.nodeSubscriptions[topic]) > 0 {
		return Error.New("Sync topic \""+topic+"\" already has a subscriber", nil)
	}
	if nodeConnection.subscribedTopics[topic] {
		return Error.New("node \""+nodeConnection.name+"\" is already subscribed to topic \""+topic+"\"", nil)
	}
	if server.nodeSubscriptions[topic] == nil {
		return Error.New("Topic \""+topic+"\" does not exist", nil)
	}
	server.nodeSubscriptions[topic][nodeConnection.name] = nodeConnection
	nodeConnection.subscribedTopics[topic] = true
	return nil
}

func (server *Server) removeSubscription(nodeConnection *nodeConnection, topic string) error {
	server.operationMutex.Lock()
	defer server.operationMutex.Unlock()
	if server.nodeSubscriptions[topic] == nil {
		return Error.New("Topic \""+topic+"\" does not exist", nil)
	}
	if !nodeConnection.subscribedTopics[topic] {
		return Error.New("node \""+nodeConnection.name+"\" is not subscribed to topic \""+topic+"\"", nil)
	}
	delete(server.nodeSubscriptions[topic], nodeConnection.name)
	delete(nodeConnection.subscribedTopics, topic)
	if len(server.nodeSubscriptions[topic]) == 0 {
		delete(server.nodeSubscriptions, topic)
	}
	return nil
}

func (server *Server) getSubscribedNodes(topic string) []*nodeConnection {
	subscribers := []*nodeConnection{}
	for _, nodeConnection := range server.nodeSubscriptions[topic] {
		subscribers = append(subscribers, nodeConnection)
	}
	return subscribers
}
