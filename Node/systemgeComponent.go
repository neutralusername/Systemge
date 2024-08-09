package Node

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

const (
	TOPIC_NODENAME          = "nodeName"
	TOPIC_RESPONSIBLETOPICS = "topics"
	TOPIC_ADDTOPIC          = "addTopic"
	TOPIC_REMOVETOPIC       = "removeTopic"
)

type systemgeComponent struct {
	config *Config.Systemge

	tcpServer *Tcp.Server

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	nodeName      string

	stopChannel                    chan bool //closing of this channel initiates the stop of the systemge component
	incomingConnectionsStopChannel chan bool //closing of this channel indicates that the incoming connection handler has stopped
	stopNode                       func()

	asyncMessageHandlerMutex sync.RWMutex
	handleAsyncMessage       func(message *Message.Message) error
	asyncMessageHandlers     map[string]AsyncMessageHandler

	syncMessageHandlerMutex sync.RWMutex
	handleSyncRequest       func(message *Message.Message) (string, error)
	syncMessageHandlers     map[string]SyncMessageHandler

	syncRequestMutex     sync.RWMutex
	syncResponseChannels map[string]*SyncResponseChannel // syncToken -> responseChannel

	outgoingConnectionMutex    sync.RWMutex
	topicResolutions           map[string]map[string]*outgoingConnection // topic -> [name -> nodeConnection]
	outgoingConnections        map[string]*outgoingConnection            // address -> nodeConnection
	outgoingConnectionAttempts map[string]*outgoingConnectionAttempt     // address -> bool

	incomingConnectionMutex sync.RWMutex
	incomingConnections     map[string]*incomingConnection // name -> nodeConnection
	handleSequentiallyMutex sync.Mutex

	// outgoing connection metrics

	outgoingConnectionRateLimiterMsgsExceeded  atomic.Uint32
	outgoingConnectionRateLimiterBytesExceeded atomic.Uint32

	outgoingConnectionAttemptsCount        atomic.Uint32
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

	receivedAddTopic    atomic.Uint32
	receivedRemoveTopic atomic.Uint32

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

	sentAddTopic    atomic.Uint32
	sentRemoveTopic atomic.Uint32

	// general metrics

	bytesReceived atomic.Uint64 // total bytes received
	bytesSent     atomic.Uint64 // total bytes sent
}

func (node *Node) startSystemgeComponent() error {
	systemge := &systemgeComponent{
		syncResponseChannels:           make(map[string]*SyncResponseChannel),
		topicResolutions:               make(map[string]map[string]*outgoingConnection),
		outgoingConnections:            make(map[string]*outgoingConnection),
		incomingConnections:            make(map[string]*incomingConnection),
		outgoingConnectionAttempts:     make(map[string]*outgoingConnectionAttempt),
		infoLogger:                     node.GetInternalInfoLogger(),
		warningLogger:                  node.GetInternalWarningLogger(),
		errorLogger:                    node.GetErrorLogger(),
		stopChannel:                    make(chan bool),
		incomingConnectionsStopChannel: make(chan bool),
		nodeName:                       node.GetName(),
		asyncMessageHandlers:           node.application.(SystemgeComponent).GetAsyncMessageHandlers(),
		syncMessageHandlers:            node.application.(SystemgeComponent).GetSyncMessageHandlers(),
		config:                         node.newNodeConfig.SystemgeConfig,
	}
	systemge.handleSyncRequest = func(message *Message.Message) (string, error) {
		systemge.syncMessageHandlerMutex.RLock()
		syncMessageHandler := systemge.syncMessageHandlers[message.GetTopic()]
		systemge.syncMessageHandlerMutex.RUnlock()
		if syncMessageHandler == nil {
			return "Not responsible for topic \"" + message.GetTopic() + "\"", Error.New("Received sync request with topic \""+message.GetTopic()+"\" for which no handler is registered", nil)
		}
		if systemge.config.HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Lock()
		}
		responsePayload, err := syncMessageHandler(node, message)
		if systemge.config.HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Unlock()
		}
		if err != nil {
			return err.Error(), Error.New("Sync message handler for topic \""+message.GetTopic()+"\" returned error", err)
		}
		return responsePayload, nil
	}
	systemge.handleAsyncMessage = func(message *Message.Message) error {
		systemge.asyncMessageHandlerMutex.RLock()
		asyncMessageHandler := systemge.asyncMessageHandlers[message.GetTopic()]
		systemge.asyncMessageHandlerMutex.RUnlock()
		if asyncMessageHandler == nil {
			return Error.New("Received async message with topic \""+message.GetTopic()+"\" for which no handler is registered", nil)
		}
		if systemge.config.HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Lock()
		}
		err := asyncMessageHandler(node, message)
		if systemge.config.HandleMessagesSequentially {
			systemge.handleSequentiallyMutex.Unlock()
		}
		if err != nil {
			return Error.New("Async message handler for topic \""+message.GetTopic()+"\" returned error", err)
		}
		return nil
	}
	systemge.stopNode = func() {
		if node.systemge == systemge {
			err := node.Stop()
			if err != nil {
				if errorLogger := systemge.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to stop node", err).Error())
				}
			}
		}
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
	go systemge.handleIncomingConnections()
	for _, endpointConfig := range systemge.config.EndpointConfigs {
		if err := systemge.attemptOutgoingConnection(endpointConfig, false); err != nil {
			return Error.New("failed to establish outgoing connection to endpoint \""+endpointConfig.Address+"\"", err)
		}
	}
	return nil
}

// stopSystemgeComponent stops the systemge component.
// blocking until all goroutines associated with the systemge component have stopped.
func (node *Node) stopSystemgeComponent() {
	systemge := node.systemge
	node.systemge = nil
	close(systemge.stopChannel)
	systemge.tcpServer.GetListener().Close()
	<-systemge.incomingConnectionsStopChannel

	systemge.outgoingConnectionMutex.Lock()
	for _, outgoingConnectionAttempt := range systemge.outgoingConnectionAttempts {
		outgoingConnectionAttempt.isAborted = true
	}
	for _, outgoingConnection := range systemge.outgoingConnections {
		outgoingConnection.netConn.Close()
		<-outgoingConnection.stopChannel
	}
	systemge.outgoingConnectionMutex.Unlock()

	systemge.incomingConnectionMutex.Lock()
	for _, incomingConnection := range systemge.incomingConnections {
		incomingConnection.netConn.Close()
		<-incomingConnection.stopChannel
	}
	systemge.incomingConnectionMutex.Unlock()
}
