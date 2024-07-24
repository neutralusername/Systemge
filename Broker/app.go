package Broker

import (
	"Systemge/Config"
	"Systemge/Node"
	"Systemge/Tcp"
	"sync"
	"sync/atomic"
)

type Broker struct {
	config *Config.Broker
	node   *Node.Node

	syncTopics  map[string]bool
	asyncTopics map[string]bool

	nodeSubscriptions map[string]map[string]*nodeConnection // topic -> [nodeName-> nodeConnection]
	nodeConnections   map[string]*nodeConnection            // nodeName -> nodeConnection
	openSyncRequests  map[string]*syncRequest               // syncKey -> syncRequest

	brokerTcpServer *Tcp.Server
	configTcpServer *Tcp.Server

	isStarted   bool
	stopChannel chan bool

	operationMutex sync.Mutex
	stateMutex     sync.Mutex

	incomingMessageCounter atomic.Uint32
	outgoingMessageCounter atomic.Uint32
	configRequestCounter   atomic.Uint32
	bytesReceivedCounter   atomic.Uint64
	bytesSentCounter       atomic.Uint64
}

func ImplementsBroker(node *Node.Node) bool {
	_, ok := node.GetApplication().(*Broker)
	return ok
}

func (broker *Broker) RetrieveIncomingMessageCounter() uint32 {
	return broker.incomingMessageCounter.Swap(0)
}

func (broker *Broker) RetrieveOutgoingMessageCounter() uint32 {
	return broker.outgoingMessageCounter.Swap(0)
}

func (broker *Broker) RetrieveConfigRequestCounter() uint32 {
	return broker.configRequestCounter.Swap(0)
}

func (broker *Broker) RetrieveBytesReceivedCounter() uint64 {
	return broker.bytesReceivedCounter.Swap(0)
}

func (broker *Broker) RetrieveBytesSentCounter() uint64 {
	return broker.bytesSentCounter.Swap(0)
}

func New(config *Config.Broker) *Broker {
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
