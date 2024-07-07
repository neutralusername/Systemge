package Broker

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Utilities"
	"net"
	"sync"
)

type Broker struct {
	logger *Utilities.Logger
	config Config.Broker

	syncTopics  map[string]bool
	asyncTopics map[string]bool

	nodeSubscriptions map[string]map[string]*nodeConnection // topic -> [nodeName-> nodeConnection]
	nodeConnections   map[string]*nodeConnection            // nodeName -> nodeConnection
	openSyncRequests  map[string]*syncRequest               // syncKey -> syncRequest

	tlsBrokerListener net.Listener
	tlsConfigListener net.Listener

	isStarted bool

	operationMutex sync.Mutex
	stateMutex     sync.Mutex
}

func New(config Config.Broker) *Broker {
	broker := &Broker{
		logger: Utilities.NewLogger(config.LoggerPath),
		config: config,

		syncTopics:  map[string]bool{},
		asyncTopics: map[string]bool{},

		nodeSubscriptions: map[string]map[string]*nodeConnection{},
		nodeConnections:   map[string]*nodeConnection{},
		openSyncRequests:  map[string]*syncRequest{},
	}
	for _, topic := range config.AsyncTopics {
		broker.addAsyncTopics(topic)
	}
	for _, topic := range config.SyncTopics {
		broker.addSyncTopics(topic)
	}
	return broker
}

func (broker *Broker) Start() error {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	if broker.isStarted {
		return Error.New("broker already started", nil)
	}
	listener, err := broker.config.Server.GetTlsListener()
	if err != nil {
		return Error.New("Failed to get listener", err)
	}
	configListener, err := broker.config.ConfigServer.GetTlsListener()
	if err != nil {
		return Error.New("Failed to get listener", err)
	}
	broker.tlsBrokerListener = listener
	broker.tlsConfigListener = configListener
	broker.isStarted = true
	go broker.handleNodeConnections()
	go broker.handleConfigConnections()

	syncTopics := []string{}
	for syncTopic := range broker.syncTopics {
		syncTopics = append(syncTopics, syncTopic)
	}
	asyncTopics := []string{}
	for asyncTopic := range broker.asyncTopics {
		asyncTopics = append(asyncTopics, asyncTopic)
	}

	if len(syncTopics) > 0 {
		err = broker.addResolverTopicsRemotely(syncTopics...)
		if err != nil {
			broker.logger.Log(Error.New("Failed to add resolver topics remotely", err).Error())
		}
	}
	if len(asyncTopics) > 0 {
		err = broker.addResolverTopicsRemotely(asyncTopics...)
		if err != nil {
			broker.logger.Log(Error.New("Failed to add resolver topic remotely", err).Error())
		}
	}
	broker.addAsyncTopics("heartbeat")
	broker.addSyncTopics("subscribe", "unsubscribe", "consume")
	return nil
}

func (broker *Broker) GetName() string {
	return broker.config.Name
}

func (broker *Broker) Stop() error {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	if !broker.isStarted {
		return Error.New("broker is not started", nil)
	}
	broker.tlsBrokerListener.Close()
	broker.tlsConfigListener.Close()
	broker.disconnectAllNodeConnections()
	broker.isStarted = false
	for syncTopic := range broker.syncTopics {
		err := broker.removeResolverTopicsRemotely(syncTopic)
		if err != nil {
			broker.logger.Log(Error.New("Failed to remove resolver topic remotely", err).Error())
		}
	}
	for asyncTopic := range broker.asyncTopics {
		err := broker.removeResolverTopicsRemotely(asyncTopic)
		if err != nil {
			broker.logger.Log(Error.New("Failed to remove resolver topic remotely", err).Error())
		}
	}
	return nil
}

func (broker *Broker) IsStarted() bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.isStarted
}
