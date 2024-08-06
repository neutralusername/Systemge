package Node

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tcp"
)

const (
	connection_nodeName_topic          = "nodeName"
	connection_responsibleTopics_topic = "topics"
)

type systemgeComponent struct {
	application              SystemgeComponent
	handleSequentiallyMutex  sync.Mutex
	asyncMessageHandlerMutex sync.Mutex
	syncMessageHandlerMutex  sync.Mutex

	config *Config.Systemge

	tcpServer *Tcp.Server

	syncRequestMutex     sync.Mutex
	syncResponseChannels map[string]*SyncResponseChannel // syncToken -> responseChannel

	outgoingConnectionMutex           sync.Mutex
	topicResolutions                  map[string]map[string]*outgoingConnection // topic -> [name -> nodeConnection]
	outgoingConnections               map[string]*outgoingConnection            // address -> nodeConnection
	currentlyInOutgoingConnectionLoop map[string]*bool                          // address -> bool

	incomingConnectionsMutex sync.Mutex
	incomingConnections      map[string]*incomingConnection // name -> nodeConnection

	// outgoing connection metrics

	outgoingConnectionRateLimiterMsgsExceeded  atomic.Uint32
	outgoingConnectionRateLimiterBytesExceeded atomic.Uint32

	outgoingConnectionAttempts             atomic.Uint32
	outgoingConnectionAttemptsSuccessful   atomic.Uint32
	outgoingConnectionAttemptsFailed       atomic.Uint32
	outgoingConnectionAttemptBytesSent     atomic.Uint64
	outgoingConnectionAttemptBytesReceived atomic.Uint64

	invalidMessagesFromOutgoingConnections atomic.Uint32

	incomingSyncResponses             atomic.Uint32
	incomingSyncSuccessResponses      atomic.Uint32
	incomingSyncFailureResponses      atomic.Uint32
	incomingSyncResponseBytesReceived atomic.Uint64

	outgoingAsyncMessages         atomic.Uint32
	outgoingAsyncMessageBytesSent atomic.Uint64

	outgoingSyncRequests         atomic.Uint32
	outgoingSyncRequestBytesSent atomic.Uint64

	// incoming connection metrics

	incomingConnectionRateLimiterMsgsExceeded  atomic.Uint32
	incomingConnectionRateLimiterBytesExceeded atomic.Uint32

	incomingConnectionAttempts             atomic.Uint32
	incomingConnectionAttemptsSuccessful   atomic.Uint32
	incomingConnectionAttemptsFailed       atomic.Uint32
	incomingConnectionAttemptBytesSent     atomic.Uint64
	incomingConnectionAttemptBytesReceived atomic.Uint64

	invalidMessagesFromIncomingConnections atomic.Uint32

	outgoingSyncResponses         atomic.Uint32
	outgoingSyncSuccessResponses  atomic.Uint32
	outgoingSyncFailureResponses  atomic.Uint32
	outgoingSyncResponseBytesSent atomic.Uint64

	incomingAsyncMessages             atomic.Uint32
	incomingAsyncMessageBytesReceived atomic.Uint64

	incomingSyncRequests             atomic.Uint32
	incomingSyncRequestBytesReceived atomic.Uint64

	// general metrics

	bytesReceived atomic.Uint64 // total bytes received
	bytesSent     atomic.Uint64 // total bytes sent
}

func (node *Node) startSystemgeComponent() error {
	if node.newNodeConfig.SystemgeConfig == nil {
		return Error.New("Systemge config missing", nil)
	}
	systemge := &systemgeComponent{
		application:                       node.application.(SystemgeComponent),
		syncResponseChannels:              make(map[string]*SyncResponseChannel),
		topicResolutions:                  make(map[string]map[string]*outgoingConnection),
		outgoingConnections:               make(map[string]*outgoingConnection),
		incomingConnections:               make(map[string]*incomingConnection),
		currentlyInOutgoingConnectionLoop: make(map[string]*bool),
		config:                            node.newNodeConfig.SystemgeConfig,
	}
	if systemge.config.TcpBufferBytes == 0 {
		systemge.config.TcpBufferBytes = 1024 * 4
	}
	tcpServer, err := Tcp.NewServer(systemge.config.ServerConfig)
	if err != nil {
		return Error.New("Failed to create tcp server", err)
	}
	systemge.tcpServer = tcpServer
	node.systemge = systemge
	go node.handleIncomingConnections()
	for _, endpointConfig := range node.systemge.config.EndpointConfigs {
		err := node.ConnectToNode(endpointConfig)
		if err != nil {
			return Error.New("Failed to establish outgoing connection to endpoint \""+endpointConfig.Address+"\"", err)
		}
	}
	return nil
}

func (node *Node) stopSystemgeComponent() error {
	systemge := node.systemge
	node.systemge = nil
	systemge.tcpServer.GetListener().Close()
	systemge.tcpServer = nil
	systemge.outgoingConnectionMutex.Lock()
	for _, brokerConnection := range systemge.outgoingConnections {
		brokerConnection.netConn.Close()
	}
	systemge.outgoingConnectionMutex.Unlock()
	systemge.incomingConnectionsMutex.Lock()
	for _, incomingConnection := range systemge.incomingConnections {
		incomingConnection.netConn.Close()
	}
	systemge.incomingConnectionsMutex.Unlock()
	return nil
}
