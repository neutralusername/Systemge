package BrokerClient

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type Client struct {
	name string

	status      int
	statusMutex sync.Mutex

	config *Config.MessageBrokerClient

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	waitGroup sync.WaitGroup

	stopChannel chan bool

	messageHandler SystemgeConnection.MessageHandler

	ongoingTopicResolutions     map[string]*resolutionAttempt
	ongoingGetBrokerConnections map[string]*getBrokerConnectionAttempt

	brokerConnections map[string]*connection            // tcpClientConfigString -> connection
	topicResolutions  map[string]map[string]*connection // topic -> [tcpClientConfigString -> connection]

	mutex sync.Mutex

	subscribedAsyncTopics map[string]bool
	subscribedSyncTopics  map[string]bool

	// metrics

	asyncMessagesSent atomic.Uint64

	syncRequestsSent      atomic.Uint64
	syncResponsesReceived atomic.Uint64

	resolutionAttempts atomic.Uint64
}

type connection struct {
	connection             SystemgeConnection.SystemgeConnection
	tcpClientConfig        *Config.TcpClient
	responsibleAsyncTopics map[string]bool
	responsibleSyncTopics  map[string]bool
}

func New(name string, config *Config.MessageBrokerClient, messageHandler SystemgeConnection.MessageHandler, dashboardCommands Commands.Handlers) *Client {
	if config == nil {
		panic(Event.New("Config is required", nil))
	}
	if config.ResolverTcpSystemgeConnectionConfig == nil {
		panic(Event.New("ResolverConnectionConfig is required", nil))
	}
	if config.ServerTcpSystemgeConnectionConfig == nil {
		panic(Event.New("ConnectionConfig is required", nil))
	}
	if len(config.ResolverTcpClientConfigs) == 0 {
		panic(Event.New("At least one ResolverTcpClientConfig is required", nil))
	}

	messageBrokerClient := &Client{
		name:                    name,
		config:                  config,
		messageHandler:          messageHandler,
		ongoingTopicResolutions: make(map[string]*resolutionAttempt),

		topicResolutions: make(map[string]map[string]*connection),

		brokerConnections: make(map[string]*connection),

		ongoingGetBrokerConnections: make(map[string]*getBrokerConnectionAttempt),

		subscribedAsyncTopics: make(map[string]bool),
		subscribedSyncTopics:  make(map[string]bool),

		status: Status.Stoped,
	}
	if config.InfoLoggerPath != "" {
		messageBrokerClient.infoLogger = Tools.NewLogger("[Info: \""+name+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		messageBrokerClient.warningLogger = Tools.NewLogger("[Warning: \""+name+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		messageBrokerClient.errorLogger = Tools.NewLogger("[Error: \""+name+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		messageBrokerClient.mailer = Tools.NewMailer(config.MailerConfig)
	}

	for _, asyncTopic := range config.AsyncTopics {
		messageBrokerClient.subscribedAsyncTopics[asyncTopic] = true
		messageBrokerClient.topicResolutions[asyncTopic] = make(map[string]*connection)
	}
	for _, syncTopic := range config.SyncTopics {
		messageBrokerClient.subscribedSyncTopics[syncTopic] = true
		messageBrokerClient.topicResolutions[syncTopic] = make(map[string]*connection)
	}
	return messageBrokerClient
}

func (messageBrokerClient *Client) Start() error {
	messageBrokerClient.statusMutex.Lock()
	defer messageBrokerClient.statusMutex.Unlock()
	if messageBrokerClient.status != Status.Stoped {
		return Event.New("Already started", nil)
	}
	messageBrokerClient.status = Status.Pending
	stopChannel := make(chan bool)
	messageBrokerClient.stopChannel = stopChannel

	for topic := range messageBrokerClient.subscribedAsyncTopics {
		messageBrokerClient.startResolutionAttempt(topic, false, stopChannel)
	}
	for topic := range messageBrokerClient.subscribedSyncTopics {
		messageBrokerClient.startResolutionAttempt(topic, true, stopChannel)
	}
	messageBrokerClient.status = Status.Started
	return nil
}

func (messageBrokerClient *Client) stop() {
	close(messageBrokerClient.stopChannel)
	messageBrokerClient.stopChannel = nil
	messageBrokerClient.waitGroup.Wait()
	messageBrokerClient.status = Status.Stoped
}
func (messageBrokerClient *Client) Stop() error {
	messageBrokerClient.statusMutex.Lock()
	defer messageBrokerClient.statusMutex.Unlock()
	if messageBrokerClient.status != Status.Started {
		return Event.New("Already started", nil)
	}
	messageBrokerClient.status = Status.Pending
	messageBrokerClient.stop()
	return nil
}

func (messageBrokerClient *Client) GetStatus() int {
	return messageBrokerClient.status
}

func (messageBrokerClient *Client) GetName() string {
	return messageBrokerClient.name
}

func getTcpClientConfigString(tcpClientConfig *Config.TcpClient) string {
	return tcpClientConfig.Address + tcpClientConfig.TlsCert
}
