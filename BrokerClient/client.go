package BrokerClient

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type MessageBrokerClient struct {
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

	dashboardClient *Dashboard.DashboardClient

	ongoingTopicResolutions map[string]*resolutionAttempt

	brokerConnections map[string]*connection            // endpointString -> connection
	topicResolutions  map[string]map[string]*connection // topic -> [endpointString -> connection]

	mutex sync.Mutex

	subscribedAsyncTopics map[string]bool
	subscribedSyncTopics  map[string]bool
}

type connection struct {
	connection             *TcpConnection.TcpConnection
	endpoint               *Config.TcpEndpoint
	responsibleAsyncTopics map[string]bool
	responsibleSyncTopics  map[string]bool
}

func New(name string, config *Config.MessageBrokerClient, systemgeMessageHandler SystemgeConnection.MessageHandler, dashboardCommands Commands.Handlers) *MessageBrokerClient {
	if config == nil {
		panic(Error.New("Config is required", nil))
	}
	if config.ResolverConnectionConfig == nil {
		panic(Error.New("ResolverConnectionConfig is required", nil))
	}
	if config.ConnectionConfig == nil {
		panic(Error.New("ConnectionConfig is required", nil))
	}
	if len(config.ResolverEndpoints) == 0 {
		panic(Error.New("At least one ResolverEndpoint is required", nil))
	}

	messageBrokerClient := &MessageBrokerClient{
		name:                    name,
		config:                  config,
		messageHandler:          systemgeMessageHandler,
		ongoingTopicResolutions: make(map[string]*resolutionAttempt),

		topicResolutions: make(map[string]map[string]*connection),

		brokerConnections: make(map[string]*connection),

		subscribedAsyncTopics: make(map[string]bool),
		subscribedSyncTopics:  make(map[string]bool),

		status: Status.STOPPED,
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

	if config.DashboardClientConfig != nil {
		messageBrokerClient.dashboardClient = Dashboard.NewClient(config.DashboardClientConfig, messageBrokerClient.Start, messageBrokerClient.Stop, messageBrokerClient.GetMetrics, messageBrokerClient.GetStatus, dashboardCommands)
		if err := messageBrokerClient.StartDashboardClient(); err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Error.New("Failed to start dashboard client", err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to start dashboard client", err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
					}
				}
			}
		}
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

func (messageBrokerClient *MessageBrokerClient) StartDashboardClient() error {
	if messageBrokerClient.dashboardClient == nil {
		return Error.New("Dashboard client is not configured", nil)
	}
	return messageBrokerClient.dashboardClient.Start()
}

func (messageBrokerClient *MessageBrokerClient) StopDashboardClient() error {
	if messageBrokerClient.dashboardClient == nil {
		return Error.New("Dashboard client is not configured", nil)
	}
	return messageBrokerClient.dashboardClient.Stop()
}

func (messageBrokerClient *MessageBrokerClient) Start() error {
	messageBrokerClient.statusMutex.Lock()
	defer messageBrokerClient.statusMutex.Unlock()
	if messageBrokerClient.status != Status.STOPPED {
		return Error.New("Already started", nil)
	}
	messageBrokerClient.status = Status.PENDING
	stopChannel := make(chan bool)
	messageBrokerClient.stopChannel = stopChannel

	for topic := range messageBrokerClient.subscribedAsyncTopics {
		err := messageBrokerClient.startResolutionAttempt(topic, false, stopChannel)
		if err != nil {
			messageBrokerClient.stop()
			return err
		}
	}
	for topic := range messageBrokerClient.subscribedSyncTopics {
		err := messageBrokerClient.startResolutionAttempt(topic, true, stopChannel)
		if err != nil {
			messageBrokerClient.stop()
			return err
		}
	}

	messageBrokerClient.status = Status.STARTED
	return nil
}

func (messageBrokerClient *MessageBrokerClient) stop() {
	close(messageBrokerClient.stopChannel)
	messageBrokerClient.stopChannel = nil
	messageBrokerClient.waitGroup.Wait()
	messageBrokerClient.status = Status.STOPPED
}
func (messageBrokerClient *MessageBrokerClient) Stop() error {
	messageBrokerClient.statusMutex.Lock()
	defer messageBrokerClient.statusMutex.Unlock()
	if messageBrokerClient.status != Status.STARTED {
		return Error.New("Already started", nil)
	}
	messageBrokerClient.status = Status.PENDING
	messageBrokerClient.stop()
	return nil
}

func (messageBrokerClient *MessageBrokerClient) GetStatus() int {
	return messageBrokerClient.status
}

func (messageBrokerClient *MessageBrokerClient) GetMetrics() map[string]uint64 {
	metrics := map[string]uint64{}
	messageBrokerClient.mutex.Lock()
	metrics["ongoingTopicResolutions"] = uint64(len(messageBrokerClient.ongoingTopicResolutions))
	metrics["brokerConnections"] = uint64(len(messageBrokerClient.brokerConnections))
	metrics["topicResolutions"] = uint64(len(messageBrokerClient.topicResolutions))
	for _, connection := range messageBrokerClient.brokerConnections {
		metrics := connection.connection.RetrieveMetrics()
		for key, value := range metrics {
			metrics[key] += value
		}
	}
	messageBrokerClient.mutex.Unlock()
	return metrics
}

func (messageBrokerClient *MessageBrokerClient) GetName() string {
	return messageBrokerClient.name
}

func getEndpointString(endpoint *Config.TcpEndpoint) string {
	return endpoint.Address + endpoint.TlsCert
}

func getASyncString(async bool) string {
	if async {
		return "async"
	}
	return "sync"
}
