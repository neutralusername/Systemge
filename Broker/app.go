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

	whitelist map[string]bool
	blacklist map[string]bool

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

		whitelist: map[string]bool{},
		blacklist: map[string]bool{},

		nodeSubscriptions: map[string]map[string]*nodeConnection{},
		nodeConnections:   map[string]*nodeConnection{},
		openSyncRequests:  map[string]*syncRequest{},
	}
	return broker
}

func (broker *Broker) isBlacklisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.blacklist[ip]
}

func (broker *Broker) isWhitelisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.whitelist[ip]
}

func (broker *Broker) validateAddress(address string) error {
	ipAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to get IP address and port from \""+address+"\"", err)
	}
	if broker.isBlacklisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to blacklist", nil)
	}
	if len(broker.whitelist) > 0 && !broker.isWhitelisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to whitelist", nil)
	}
	return nil
}
