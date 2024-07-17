package Broker

import (
	"Systemge/Error"
	"Systemge/Node"
)

func (broker *Broker) OnStart(node *Node.Node) error {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	if broker.isStarted {
		return Error.New("Broker \""+node.GetName()+"\" is already started", nil)
	}
	listener, err := broker.config.Server.GetListener()
	if err != nil {
		return Error.New("Failed to get listener for broker \""+node.GetName()+"\"", err)
	}
	configListener, err := broker.config.ConfigServer.GetListener()
	if err != nil {
		return Error.New("Failed to get config listener for broker \""+node.GetName()+"\"", err)
	}
	broker.tlsBrokerListener = listener
	broker.tlsConfigListener = configListener
	broker.isStarted = true
	broker.node = node
	broker.stopChannel = make(chan bool)

	topicsToAddToResolver := []string{}
	for _, topic := range broker.config.AsyncTopics {
		broker.addAsyncTopics(topic)
		topicsToAddToResolver = append(topicsToAddToResolver, topic)
	}
	for _, topic := range broker.config.SyncTopics {
		broker.addSyncTopics(topic)
		topicsToAddToResolver = append(topicsToAddToResolver, topic)
	}
	if len(topicsToAddToResolver) > 0 {
		err = broker.addResolverTopicsRemotely(topicsToAddToResolver...)
		if err != nil {
			broker.stop(node, false)
			return Error.New("Failed to add resolver topics remotely on broker \""+node.GetName()+"\"", err)
		}
	}
	broker.addAsyncTopics("heartbeat")
	broker.addSyncTopics("subscribe", "unsubscribe")
	go broker.handleNodeConnections()
	go broker.handleConfigConnections()
	return nil
}

func (broker *Broker) OnStop(node *Node.Node) error {
	return broker.stop(node, true)
}

func (broker *Broker) stop(node *Node.Node, lock bool) error {
	if lock {
		broker.stateMutex.Lock()
		defer broker.stateMutex.Unlock()
	}
	if !broker.isStarted {
		return Error.New("Broker \""+node.GetName()+"\" is already stopped", nil)
	}
	broker.tlsBrokerListener.Close()
	broker.tlsBrokerListener = nil
	broker.tlsConfigListener.Close()
	broker.tlsConfigListener = nil
	broker.removeAllNodeConnections()
	broker.node = nil
	broker.isStarted = false
	close(broker.stopChannel)
	topics := []string{}
	for syncTopic := range broker.syncTopics {
		if syncTopic != "subscribe" && syncTopic != "unsubscribe" {
			topics = append(topics, syncTopic)
		}
		delete(broker.syncTopics, syncTopic)
	}
	for asyncTopic := range broker.asyncTopics {
		if asyncTopic != "heartbeat" {
			topics = append(topics, asyncTopic)
		}
		delete(broker.asyncTopics, asyncTopic)
	}
	if len(topics) > 0 {
		err := broker.removeResolverTopicsRemotely(topics...)
		if err != nil { // This should never happen
			node.GetLogger().Error(Error.New("Failed to remove resolver topics remotely on broker \""+node.GetName()+"\"", err).Error())
		} else {
			node.GetLogger().Info("Removed resolver topics remotely on broker \"" + node.GetName() + "\"")
		}
	}
	node.GetLogger().Info(Error.New("Stopped broker \""+node.GetName()+"\"", nil).Error())
	return nil
}
