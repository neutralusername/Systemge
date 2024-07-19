package Broker

import (
	"Systemge/Config"
	"Systemge/Error"
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
	return broker
}

func (broker *Broker) isBrokerBlacklisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.brokerBlacklist[ip]
}

func (broker *Broker) isBrokerWhitelisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.brokerWhitelist[ip]
}

func (broker *Broker) validateAddressBroker(address string) error {
	ipAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to get IP address and port from \""+address+"\"", err)
	}
	if broker.isBrokerBlacklisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to blacklist", nil)
	}
	if len(broker.brokerWhitelist) > 0 && !broker.isBrokerWhitelisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to whitelist", nil)
	}
	return nil
}

func (broker *Broker) isConfigBlacklisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.configBlacklist[ip]
}

func (broker *Broker) isConfigWhitelisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.configWhitelist[ip]
}

func (broker *Broker) validateAddressConfig(address string) error {
	ipAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to get IP address and port from \""+address+"\"", err)
	}
	if broker.isConfigBlacklisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to blacklist", nil)
	}
	if len(broker.configWhitelist) > 0 && !broker.isConfigWhitelisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to whitelist", nil)
	}
	return nil
}
