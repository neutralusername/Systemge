package Broker

import (
	"Systemge/Config"
	"Systemge/Node"
	"net"
	"sync"
)

type Broker struct {
	config *Config.Broker
	node   *Node.Node

	syncTopics  map[string]bool
	asyncTopics map[string]bool

	brokerWhitelist map[string]bool
	brokerBlacklist map[string]bool
	configWhitelist map[string]bool
	configBlacklist map[string]bool

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

func New(config *Config.Broker) *Broker {
	broker := &Broker{
		config: config,

		syncTopics:  map[string]bool{},
		asyncTopics: map[string]bool{},

		brokerWhitelist: map[string]bool{},
		brokerBlacklist: map[string]bool{},
		configWhitelist: map[string]bool{},
		configBlacklist: map[string]bool{},

		nodeSubscriptions: map[string]map[string]*nodeConnection{},
		nodeConnections:   map[string]*nodeConnection{},
		openSyncRequests:  map[string]*syncRequest{},
	}
	for _, ip := range broker.config.BrokerWhitelist {
		broker.brokerWhitelist[ip] = true
	}
	for _, ip := range broker.config.BrokerBlacklist {
		broker.brokerBlacklist[ip] = true
	}
	for _, ip := range broker.config.ConfigWhitelist {
		broker.configWhitelist[ip] = true
	}
	for _, ip := range broker.config.ConfigBlacklist {
		broker.configBlacklist[ip] = true
	}
	return broker
}
