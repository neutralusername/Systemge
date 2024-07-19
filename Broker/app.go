package Broker

import (
	"Systemge/Config"
	"Systemge/Node"
	"Systemge/Tools"
	"net"
	"sync"
)

type Broker struct {
	config *Config.Broker
	node   *Node.Node

	syncTopics  map[string]bool
	asyncTopics map[string]bool

	brokerWhitelist Tools.AccessControlList
	brokerBlacklist Tools.AccessControlList
	configWhitelist Tools.AccessControlList
	configBlacklist Tools.AccessControlList

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

		brokerWhitelist: Tools.AccessControlList{},
		brokerBlacklist: Tools.AccessControlList{},
		configWhitelist: Tools.AccessControlList{},
		configBlacklist: Tools.AccessControlList{},

		nodeSubscriptions: map[string]map[string]*nodeConnection{},
		nodeConnections:   map[string]*nodeConnection{},
		openSyncRequests:  map[string]*syncRequest{},
	}
	for _, ip := range broker.config.BrokerWhitelist {
		broker.brokerWhitelist.Add(ip)
	}
	for _, ip := range broker.config.BrokerBlacklist {
		broker.brokerBlacklist.Add(ip)
	}
	for _, ip := range broker.config.ConfigWhitelist {
		broker.configWhitelist.Add(ip)
	}
	for _, ip := range broker.config.ConfigBlacklist {
		broker.configBlacklist.Add(ip)
	}
	return broker
}
