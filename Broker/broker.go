package Broker

import (
	"Systemge/Config"
	"Systemge/Error"
	"net"
	"sync"
)

type Broker struct {
	config Config.Broker

	syncTopics  map[string]bool
	asyncTopics map[string]bool

	nodeSubscriptions map[string]map[string]*nodeConnection // topic -> [nodeName-> nodeConnection]
	nodeConnections   map[string]*nodeConnection            // nodeName -> nodeConnection
	openSyncRequests  map[string]*syncRequest               // syncKey -> syncRequest

	tlsBrokerListener net.Listener
	tlsConfigListener net.Listener

	isStarted   bool
	stopChannel chan bool

	operationMutex sync.Mutex
	stateMutex     sync.Mutex
}

func New(config Config.Broker) *Broker {
	broker := &Broker{
		config: config,

		syncTopics:  map[string]bool{},
		asyncTopics: map[string]bool{},

		nodeSubscriptions: map[string]map[string]*nodeConnection{},
		nodeConnections:   map[string]*nodeConnection{},
		openSyncRequests:  map[string]*syncRequest{},
	}
	return broker
}

func (broker *Broker) Start() error {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	if broker.isStarted {
		return Error.New("Broker \""+broker.GetName()+"\" is already started", nil)
	}
	listener, err := broker.config.Server.GetTlsListener()
	if err != nil {
		return Error.New("Failed to get listener for broker \""+broker.GetName()+"\"", err)
	}
	configListener, err := broker.config.ConfigServer.GetTlsListener()
	if err != nil {
		return Error.New("Failed to get config listener for broker \""+broker.GetName()+"\"", err)
	}
	broker.tlsBrokerListener = listener
	broker.tlsConfigListener = configListener
	broker.isStarted = true
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
			broker.stop(false)
			return Error.New("Failed to add resolver topics remotely on broker \""+broker.GetName()+"\"", err)
		}
	}
	broker.addAsyncTopics("heartbeat")
	broker.addSyncTopics("subscribe", "unsubscribe")
	go broker.handleNodeConnections()
	go broker.handleConfigConnections()
	broker.config.Logger.Info(Error.New("Started broker \""+broker.GetName()+"\"", nil).Error())
	return nil
}

func (broker *Broker) GetName() string {
	return broker.config.Name
}

func (broker *Broker) Stop() error {
	return broker.stop(true)
}

func (broker *Broker) stop(lock bool) error {
	if lock {
		broker.stateMutex.Lock()
		defer broker.stateMutex.Unlock()
	}
	if !broker.isStarted {
		return Error.New("Broker \""+broker.GetName()+"\" is already stopped", nil)
	}
	broker.tlsBrokerListener.Close()
	broker.tlsBrokerListener = nil
	broker.tlsConfigListener.Close()
	broker.tlsConfigListener = nil
	broker.disconnectAllNodeConnections()
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
		if err != nil {
			broker.config.Logger.Error(Error.New("Failed to remove resolver topics remotely on broker \""+broker.GetName()+"\"", err).Error())
		} else {
			broker.config.Logger.Info("Removed resolver topics remotely on broker \"" + broker.GetName() + "\"")
		}
	}
	broker.config.Logger.Info(Error.New("Stopped broker \""+broker.GetName()+"\"", nil).Error())
	return nil
}

func (broker *Broker) IsStarted() bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.isStarted
}
