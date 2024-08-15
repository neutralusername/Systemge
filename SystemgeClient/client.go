package SystemgeClient

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeClient struct {
	status      int
	startMutex  sync.Mutex
	stopMutex   sync.Mutex
	statusMutex sync.Mutex

	config *Config.SystemgeConnection

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
	randomizer    *Tools.Randomizer

	stopChannel          chan bool
	syncResponseChannels map[string]chan *Message.Message
	serverConnection     *serverConnection
	mutex                sync.RWMutex

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

func New(config *Config.SystemgeConnection) *SystemgeClient {
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

		syncResponseChannels: make(map[string]chan *Message.Message),
	}
}

func (client *SystemgeClient) Start() error {
	client.startMutex.Lock()
	defer client.startMutex.Unlock()

	client.statusMutex.Lock()
	if client.status != Status.STOPPED {
		client.statusMutex.Unlock()
		return Error.New("SystemgeClient is not in stopped state", nil)
	}
	client.status = Status.PENDING
	client.statusMutex.Unlock()

	serverConnection, err := client.attemptServerConnection(client.config.EndpointConfig)
	if err != nil {
		client.status = Status.STOPPED
		return Error.New("failed to establish server connection to endpoint \""+client.config.EndpointConfig.Address+"\"", err)
	}

	client.statusMutex.Lock()
	if client.status != Status.PENDING {
		client.statusMutex.Unlock()
		serverConnection.netConn.Close()
		return Error.New("SystemgeClient stopped during startup", nil)
	}
	client.status = Status.STARTED
	client.stopChannel = make(chan bool)
	go client.handleServerConnectionMessages(serverConnection)
	client.statusMutex.Unlock()
	return nil
}

func (client *SystemgeClient) Stop() error {
	client.stopMutex.Lock()
	defer client.stopMutex.Unlock()

	client.statusMutex.Lock()
	if client.status == Status.STOPPED {
		client.statusMutex.Unlock()
		return Error.New("SystemgeClient is already in stopped state", nil)
	}
	client.status = Status.PENDING
	client.statusMutex.Unlock()

	close(client.stopChannel)
	client.serverConnection.netConn.Close()
	<-client.serverConnection.stopChannel
	client.serverConnection = nil

	client.statusMutex.Lock()
	client.status = Status.STOPPED
	client.statusMutex.Unlock()
	return nil
}

func (client *SystemgeClient) GetName() string {
	return client.config.Name
}

func (client *SystemgeClient) GetStatus() int {
	return client.status
}
