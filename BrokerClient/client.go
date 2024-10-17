package BrokerClient

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/Commands"
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/SystemgeConnection"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

type Client struct {
	name string

	status      int
	statusMutex sync.Mutex

	config *configs.MessageBrokerClient

	infoLogger    *tools.Logger
	warningLogger *tools.Logger
	errorLogger   *tools.Logger
	mailer        *tools.Mailer

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
	tcpClientConfig        *configs.TcpClient
	responsibleAsyncTopics map[string]bool
	responsibleSyncTopics  map[string]bool
}

func New(name string, config *configs.MessageBrokerClient, messageHandler SystemgeConnection.MessageHandler, dashboardCommands Commands.Handlers) *Client {
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

		status: status.Stopped,
	}
	if config.InfoLoggerPath != "" {
		messageBrokerClient.infoLogger = tools.NewLogger("[Info: \""+name+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		messageBrokerClient.warningLogger = tools.NewLogger("[Warning: \""+name+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		messageBrokerClient.errorLogger = tools.NewLogger("[Error: \""+name+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		messageBrokerClient.mailer = tools.NewMailer(config.MailerConfig)
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
	if messageBrokerClient.status != status.Stopped {
		return Event.New("Already started", nil)
	}
	messageBrokerClient.status = status.Pending
	stopChannel := make(chan bool)
	messageBrokerClient.stopChannel = stopChannel

	for topic := range messageBrokerClient.subscribedAsyncTopics {
		messageBrokerClient.startResolutionAttempt(topic, false, stopChannel)
	}
	for topic := range messageBrokerClient.subscribedSyncTopics {
		messageBrokerClient.startResolutionAttempt(topic, true, stopChannel)
	}
	messageBrokerClient.status = status.Started
	return nil
}

func (messageBrokerClient *Client) stop() {
	close(messageBrokerClient.stopChannel)
	messageBrokerClient.stopChannel = nil
	messageBrokerClient.waitGroup.Wait()
	messageBrokerClient.status = status.Stopped
}
func (messageBrokerClient *Client) Stop() error {
	messageBrokerClient.statusMutex.Lock()
	defer messageBrokerClient.statusMutex.Unlock()
	if messageBrokerClient.status != status.Started {
		return Event.New("Already started", nil)
	}
	messageBrokerClient.status = status.Pending
	messageBrokerClient.stop()
	return nil
}

func (messageBrokerClient *Client) GetStatus() int {
	return messageBrokerClient.status
}

func (messageBrokerClient *Client) GetName() string {
	return messageBrokerClient.name
}

func getTcpClientConfigString(tcpClientConfig *configs.TcpClient) string {
	return tcpClientConfig.Address + tcpClientConfig.TlsCert
}
