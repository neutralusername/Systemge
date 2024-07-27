package Broker

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Node"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

func (broker *Broker) OnStart(node *Node.Node) error {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	if broker.isStarted {
		return Error.New("Broker \""+node.GetName()+"\" is already started", nil)
	}
	listener, err := Tcp.NewServer(broker.config.Server)
	if err != nil {
		return Error.New("Failed to get listener for broker \""+node.GetName()+"\"", err)
	}
	configListener, err := Tcp.NewServer(broker.config.ConfigServer)
	if err != nil {
		listener.GetListener().Close()
		return Error.New("Failed to get config listener for broker \""+node.GetName()+"\"", err)
	}
	broker.brokerTcpServer = listener
	broker.configTcpServer = configListener
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
			return err
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
	broker.isStarted = false
	broker.brokerTcpServer.GetListener().Close()
	broker.brokerTcpServer = nil
	broker.configTcpServer.GetListener().Close()
	broker.configTcpServer = nil
	broker.removeAllNodeConnections()
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
		if err := broker.removeResolverTopicsRemotely(topics...); err != nil {
			if errorLogger := node.GetErrorLogger(); errorLogger != nil {
				errorLogger.Log(Error.New("Failed to remove resolver topics remotely", err).Error())
				if mailer := node.GetMailer(); mailer != nil {
					mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to remove resolver topics remotely", err).Error()))
				}
			}
		} else {
			if infoLogger := node.GetInfoLogger(); infoLogger != nil {
				infoLogger.Log("Removed resolver topics remotely")
			}
		}
	}
	return nil
}
