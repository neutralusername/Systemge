package Broker

import "Systemge/Error"

func (broker *Broker) addSubscription(nodeConnection *nodeConnection, topic string) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	if !broker.asyncTopics[topic] && !broker.syncTopics[topic] {
		return Error.New("Topic \""+topic+"\" does not exist on broker \""+broker.GetName()+"\"", nil)
	}
	if broker.syncTopics[topic] && len(broker.nodeSubscriptions[topic]) > 0 {
		return Error.New("Sync topic \""+topic+"\" already has a subscriber", nil)
	}
	if nodeConnection.subscribedTopics[topic] {
		return Error.New("node \""+nodeConnection.name+"\" is already subscribed to topic \""+topic+"\"", nil)
	}
	if broker.nodeSubscriptions[topic] == nil {
		return Error.New("Topic \""+topic+"\" does not exist", nil)
	}
	broker.nodeSubscriptions[topic][nodeConnection.name] = nodeConnection
	nodeConnection.subscribedTopics[topic] = true
	return nil
}

func (broker *Broker) removeSubscription(nodeConnection *nodeConnection, topic string) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	if broker.nodeSubscriptions[topic] == nil {
		return Error.New("Topic \""+topic+"\" does not exist", nil)
	}
	if !nodeConnection.subscribedTopics[topic] {
		return Error.New("node \""+nodeConnection.name+"\" is not subscribed to topic \""+topic+"\"", nil)
	}
	delete(broker.nodeSubscriptions[topic], nodeConnection.name)
	delete(nodeConnection.subscribedTopics, topic)
	if len(broker.nodeSubscriptions[topic]) == 0 {
		delete(broker.nodeSubscriptions, topic)
	}
	return nil
}

func (broker *Broker) getSubscribedNodes(topic string) []*nodeConnection {
	subscribers := []*nodeConnection{}
	for _, nodeConnection := range broker.nodeSubscriptions[topic] {
		subscribers = append(subscribers, nodeConnection)
	}
	return subscribers
}
