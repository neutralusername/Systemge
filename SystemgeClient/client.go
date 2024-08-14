package Node

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeClient struct {
	config *Config.SystemgeClient

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
	randomizer    *Tools.Randomizer

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

func New(config *Config.SystemgeClient) *SystemgeClient {
	if config == nil {
		panic("SystemgeClient config is nil")
	}
	return &SystemgeClient{
		config:        config,
		infoLogger:    Tools.NewLogger("[Info: \""+config.Name+"\"] ", config.InfoLoggerPath),
		warningLogger: Tools.NewLogger("[Warning: \""+config.Name+"\"] ", config.WarningLoggerPath),
		errorLogger:   Tools.NewLogger("[Error: \""+config.Name+"\"] ", config.ErrorLoggerPath),
		mailer:        Tools.NewMailer(config.MailerConfig),
		randomizer:    Tools.NewRandomizer(config.RandomizerSeed),

		syncResponseChannels:       make(map[string]*SyncResponseChannel),
		topicResolutions:           make(map[string]map[string]*outgoingConnection),
		outgoingConnections:        make(map[string]*outgoingConnection),
		outgoingConnectionAttempts: make(map[string]*outgoingConnectionAttempt),
		stopChannel:                make(chan bool),
	}
}

func (client *SystemgeClient) Start() error {
	for _, endpointConfig := range client.config.EndpointConfigs {
		if err := client.attemptOutgoingConnection(endpointConfig, false); err != nil {
			return Error.New("failed to establish outgoing connection to endpoint \""+endpointConfig.Address+"\"", err)
		}
	}
	return nil
}

func (client *SystemgeClient) Stop() error {
	close(client.stopChannel)

	client.outgoingConnectionMutex.Lock()
	for _, outgoingConnectionAttempt := range client.outgoingConnectionAttempts {
		outgoingConnectionAttempt.isAborted = true
	}
	for _, outgoingConnection := range client.outgoingConnections {
		outgoingConnection.netConn.Close()
		<-outgoingConnection.stopChannel
	}
	client.outgoingConnectionMutex.Unlock()
	return nil
}

func (systemge *SystemgeClient) validateMessage(message *Message.Message) error {
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
