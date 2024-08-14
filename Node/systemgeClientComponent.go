package Node

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type systemgeClientComponent struct {
	config *Config.SystemgeClient

	nodeName      string
	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger

	stopNode    func()
	stopChannel chan bool //closing of this channel initiates the stop of the systemge component

	syncRequestMutex     sync.RWMutex
	syncResponseChannels map[string]*SyncResponseChannel // syncToken -> responseChannel

	outgoingConnectionMutex    sync.RWMutex
	topicResolutions           map[string]map[string]*outgoingConnection // topic -> [name -> nodeConnection]
	outgoingConnections        map[string]*outgoingConnection            // address -> nodeConnection
	outgoingConnectionAttempts map[string]*outgoingConnectionAttempt     // address -> bool

	// outgoing connection metrics

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

	syncSuccessResponsesReceived atomic.Uint32
	syncFailureResponsesReceived atomic.Uint32
	syncResponseBytesReceived    atomic.Uint64

	asyncMessagesSent     atomic.Uint32
	asyncMessageBytesSent atomic.Uint64

	syncRequestsSent     atomic.Uint32
	syncRequestBytesSent atomic.Uint64

	topicAddReceived    atomic.Uint32
	topicRemoveReceived atomic.Uint32
}

func (node *Node) startSystemgeClientComponent() error {
	systemgeClient := &systemgeClientComponent{
		config:                     node.config.SystemgeClientConfig,
		nodeName:                   node.GetName(),
		infoLogger:                 node.infoLogger,
		warningLogger:              node.warningLogger,
		errorLogger:                node.errorLogger,
		syncResponseChannels:       make(map[string]*SyncResponseChannel),
		topicResolutions:           make(map[string]map[string]*outgoingConnection),
		outgoingConnections:        make(map[string]*outgoingConnection),
		outgoingConnectionAttempts: make(map[string]*outgoingConnectionAttempt),
		stopChannel:                make(chan bool),
	}
	systemgeClient.stopNode = func() {
		if node.systemgeClient == systemgeClient {
			err := node.Stop()
			if err != nil {
				if errorLogger := systemgeClient.errorLogger; errorLogger != nil {
					errorLogger.Log(Error.New("Failed to stop node", err).Error())
				}
			}
		}
	}
	node.systemgeClient = systemgeClient
	for _, endpointConfig := range systemgeClient.config.EndpointConfigs {
		if err := systemgeClient.attemptOutgoingConnection(endpointConfig, false); err != nil {
			return Error.New("failed to establish outgoing connection to endpoint \""+endpointConfig.Address+"\"", err)
		}
	}
	return nil
}

func (node *Node) stopSystemgeClientComponent() {
	systemge := node.systemgeClient
	node.systemgeClient = nil
	close(systemge.stopChannel)

	systemge.outgoingConnectionMutex.Lock()
	for _, outgoingConnectionAttempt := range systemge.outgoingConnectionAttempts {
		outgoingConnectionAttempt.isAborted = true
	}
	for _, outgoingConnection := range systemge.outgoingConnections {
		outgoingConnection.netConn.Close()
		<-outgoingConnection.stopChannel
	}
	systemge.outgoingConnectionMutex.Unlock()
}

func (systemge *systemgeClientComponent) validateMessage(message *Message.Message) error {
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
