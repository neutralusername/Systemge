package Node

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
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

	syncResponseChannels     map[string]*SyncResponseChannel         // syncToken -> responseChannel
	topicResolutions         map[string]map[string]*serverConnection // topic -> [name -> nodeConnection]
	serverConnections        map[string]*serverConnection            // address -> nodeConnection
	serverConnectionAttempts map[string]*serverConnectionAttempt     // address -> bool

	serverConnectionMutex sync.RWMutex
	syncRequestMutex      sync.RWMutex

	// metrics

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

		syncResponseChannels:     make(map[string]*SyncResponseChannel),
		topicResolutions:         make(map[string]map[string]*serverConnection),
		serverConnections:        make(map[string]*serverConnection),
		serverConnectionAttempts: make(map[string]*serverConnectionAttempt),
		stopChannel:              make(chan bool),
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

	client.serverConnectionMutex.Lock()
	for _, outgoingConnectionAttempt := range client.serverConnectionAttempts {
		outgoingConnectionAttempt.isAborted = true
	}
	for _, outgoingConnection := range client.serverConnections {
		outgoingConnection.netConn.Close()
		<-outgoingConnection.stopChannel
	}
	client.serverConnectionMutex.Unlock()
	return nil
}

func (client *SystemgeClient) GetName() string {
	return client.config.Name
}
