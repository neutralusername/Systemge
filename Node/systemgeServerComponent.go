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

type systemgeServerComponent struct {
	config *Config.SystemgeServer

	tcpServer *Tcp.Server

	nodeName      string
	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger

	ipRateLimiter *Tools.IpRateLimiter

	stopChannel                          chan bool //closing of this channel initiates the stop of the systemge component
	incomingConnectionsStopChannel       chan bool //closing of this channel indicates that the incoming connection handler has stopped
	allIncomingConnectionsStoppedChannel chan bool //closing of this channel indicates that all incoming connections have stopped
	messageHandlerChannel                chan func()

	asyncMessageHandlerMutex sync.RWMutex
	handleAsyncMessage       func(message *Message.Message) error
	asyncMessageHandlers     map[string]AsyncMessageHandler

	syncMessageHandlerMutex sync.RWMutex
	handleSyncRequest       func(message *Message.Message) (string, error)
	syncMessageHandlers     map[string]SyncMessageHandler

	incomingConnectionMutex sync.RWMutex
	incomingConnections     map[string]*incomingConnection // name -> nodeConnection

	// incoming connection metrics
	bytesReceived           atomic.Uint64
	bytesSent               atomic.Uint64
	invalidMessagesReceived atomic.Uint32

	messageRateLimiterExceeded atomic.Uint32
	byteRateLimiterExceeded    atomic.Uint32

	connectionAttempts             atomic.Uint32
	connectionAttemptsSuccessful   atomic.Uint32
	connectionAttemptsFailed       atomic.Uint32
	connectionAttemptBytesSent     atomic.Uint64
	connectionAttemptBytesReceived atomic.Uint64

	syncSuccessResponsesSent atomic.Uint32
	syncFailureResponsesSent atomic.Uint32
	syncResponseBytesSent    atomic.Uint64

	asyncMessagesReceived     atomic.Uint32
	asyncMessageBytesReceived atomic.Uint64

	syncRequestsReceived     atomic.Uint32
	syncRequestBytesReceived atomic.Uint64

	topicAddSent    atomic.Uint32
	topicRemoveSent atomic.Uint32
}

func (node *Node) startSystemgeServerComponent() error {
	systemgeServer := &systemgeServerComponent{
		incomingConnections:                  make(map[string]*incomingConnection),
		infoLogger:                           node.GetInternalInfoLogger(),
		warningLogger:                        node.GetInternalWarningLogger(),
		errorLogger:                          node.GetErrorLogger(),
		stopChannel:                          make(chan bool),
		incomingConnectionsStopChannel:       make(chan bool),
		allIncomingConnectionsStoppedChannel: make(chan bool),
		nodeName:                             node.GetName(),
		asyncMessageHandlers:                 node.application.(SystemgeServerComponent).GetAsyncMessageHandlers(),
		syncMessageHandlers:                  node.application.(SystemgeServerComponent).GetSyncMessageHandlers(),
		config:                               node.newNodeConfig.SystemgeServerConfig,
	}
	if systemgeServer.config.TcpBufferBytes == 0 {
		systemgeServer.config.TcpBufferBytes = 1024 * 4
	}
	tcpServer, err := Tcp.NewServer(systemgeServer.config.ServerConfig)
	if err != nil {
		return Error.New("Failed to create tcp server", err)
	}
	systemgeServer.tcpServer = tcpServer
	node.systemgeServer = systemgeServer
	systemgeServer.handleSyncRequest = func(message *Message.Message) (string, error) {
		systemgeServer.syncMessageHandlerMutex.RLock()
		syncMessageHandler := systemgeServer.syncMessageHandlers[message.GetTopic()]
		systemgeServer.syncMessageHandlerMutex.RUnlock()
		if syncMessageHandler == nil {
			return "Not responsible for topic \"" + message.GetTopic() + "\"", Error.New("Received sync request with topic \""+message.GetTopic()+"\" for which no handler is registered", nil)
		}
		responsePayload, err := syncMessageHandler(node, message)
		if err != nil {
			return err.Error(), Error.New("Sync message handler for topic \""+message.GetTopic()+"\" returned error", err)
		}
		return responsePayload, nil
	}
	systemgeServer.handleAsyncMessage = func(message *Message.Message) error {
		systemgeServer.asyncMessageHandlerMutex.RLock()
		asyncMessageHandler := systemgeServer.asyncMessageHandlers[message.GetTopic()]
		systemgeServer.asyncMessageHandlerMutex.RUnlock()
		if asyncMessageHandler == nil {
			return Error.New("Received async message with topic \""+message.GetTopic()+"\" for which no handler is registered", nil)
		}
		err := asyncMessageHandler(node, message)
		if err != nil {
			return Error.New("Async message handler for topic \""+message.GetTopic()+"\" returned error", err)
		}
		return nil
	}

	if systemgeServer.config.IpRateLimiter != nil {
		systemgeServer.ipRateLimiter = Tools.NewIpRateLimiter(systemgeServer.config.IpRateLimiter)
	}
	if systemgeServer.config.ProcessAllMessagesSequentially {
		systemgeServer.messageHandlerChannel = make(chan func(), systemgeServer.config.ProcessAllMessagesSequentiallyChannelSize)
		if systemgeServer.config.ProcessAllMessagesSequentiallyChannelSize == 0 {
			go func() {
				for {
					select {
					case f := <-systemgeServer.messageHandlerChannel:
						f()
					case <-systemgeServer.stopChannel:
						return
					}
				}
			}()
		} else {
			go func() {
				for {
					select {
					case f := <-systemgeServer.messageHandlerChannel:
						if systemgeServer.errorLogger != nil && len(systemgeServer.messageHandlerChannel) >= systemgeServer.config.ProcessAllMessagesSequentiallyChannelSize-1 {
							systemgeServer.errorLogger.Log("ProcessAllMessagesSequentiallyChannelSize reached (increase ProcessAllMessagesSequentiallyChannelSize otherwise message order of arrival is not guaranteed)")
						}
						f()
					case <-systemgeServer.allIncomingConnectionsStoppedChannel:
						return
					}
				}
			}()
		}

	}
	go systemgeServer.handleIncomingConnections()
	return nil
}

// stopSystemgeServerComponent stops the systemge component.
// blocking until all goroutines associated with the systemge component have stopped.
func (node *Node) stopSystemgeServerComponent() {
	systemge := node.systemgeServer
	node.systemgeServer = nil
	close(systemge.stopChannel)

	if systemge.ipRateLimiter != nil {
		systemge.ipRateLimiter.Stop()
	}

	systemge.tcpServer.GetListener().Close()
	<-systemge.incomingConnectionsStopChannel

	systemge.incomingConnectionMutex.Lock()
	for _, incomingConnection := range systemge.incomingConnections {
		incomingConnection.netConn.Close()
		<-incomingConnection.stopChannel
	}
	systemge.incomingConnectionMutex.Unlock()
	close(systemge.allIncomingConnectionsStoppedChannel)
}

func (systemge *systemgeServerComponent) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := systemge.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncTokenToken()) > maxSyncTokenSize {
		return Error.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Error.New("Message missing topic", nil)
	}
	if maxTopicSize := systemge.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Error.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := systemge.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Error.New("Message payload exceeds maximum size", nil)
	}
	return nil
}
