package BrokerClient

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
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

	dashboardClient *Dashboard.DashboardClient

	ongoingTopicResolutions     map[string]*resolutionAttempt
	ongoingGetBrokerConnections map[string]*getBrokerConnectionAttempt

	brokerConnections map[string]*connection            // endpointString -> connection
	topicResolutions  map[string]map[string]*connection // topic -> [endpointString -> connection]

	mutex sync.Mutex

	subscribedAsyncTopics map[string]bool
	subscribedSyncTopics  map[string]bool
}

type connection struct {
	connection             SystemgeConnection.SystemgeConnection
	endpoint               *Config.TcpClient
	responsibleAsyncTopics map[string]bool
	responsibleSyncTopics  map[string]bool
}

func New(name string, config *Config.MessageBrokerClient, systemgeMessageHandler SystemgeConnection.MessageHandler, dashboardCommands Commands.Handlers) *Client {
	if config == nil {
		panic(Error.New("Config is required", nil))
	}
	if config.ResolverConnectionConfig == nil {
		panic(Error.New("ResolverConnectionConfig is required", nil))
	}
	if config.ConnectionConfig == nil {
		panic(Error.New("ConnectionConfig is required", nil))
	}
	if len(config.ResolverClientConfigs) == 0 {
		panic(Error.New("At least one ResolverEndpoint is required", nil))
	}

	messageBrokerClient := &Client{
		name:                    name,
		config:                  config,
		messageHandler:          systemgeMessageHandler,
		ongoingTopicResolutions: make(map[string]*resolutionAttempt),

		topicResolutions: make(map[string]map[string]*connection),

		brokerConnections: make(map[string]*connection),

		ongoingGetBrokerConnections: make(map[string]*getBrokerConnectionAttempt),

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
		dashboardCommandHandler := Commands.Handlers{}
		dashboardCommandHandler["resolveTopic"] = func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.ResolveTopic(args[0])
		}
		dashboardCommandHandler["resolveSubscribeTopic"] = func(args []string) (string, error) {
			return "success", messageBrokerClient.ResolveSubscribeTopics()
		}
		dashboardCommandHandler["getAsyncSubscribeTopics"] = func(args []string) (string, error) {
			topics := messageBrokerClient.GetAsyncSubscribeTopics()
			return Helpers.JsonMarshal(topics), nil
		}
		dashboardCommandHandler["getSyncSubscribeTopics"] = func(args []string) (string, error) {
			topics := messageBrokerClient.GetSyncSubscribeTopics()
			return Helpers.JsonMarshal(topics), nil
		}
		dashboardCommandHandler["addAsyncSubscribeTopic"] = func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.AddAsyncSubscribeTopic(args[0])
		}
		dashboardCommandHandler["addSyncSubscribeTopic"] = func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.AddSyncSubscribeTopic(args[0])
		}
		dashboardCommandHandler["removeAsyncSubscribeTopic"] = func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.RemoveAsyncSubscribeTopic(args[0])
		}
		dashboardCommandHandler["removeSyncSubscribeTopic"] = func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.RemoveSyncSubscribeTopic(args[0])
		}
		for key, value := range dashboardCommands {
			// may override the default commands
			dashboardCommandHandler[key] = value
		}
		messageBrokerClient.dashboardClient = Dashboard.NewClient(name+"_dashboardClient", config.DashboardClientConfig, messageBrokerClient.Start, messageBrokerClient.Stop, messageBrokerClient.GetMetrics, messageBrokerClient.GetStatus, dashboardCommandHandler)
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

func (messageBrokerClient *Client) StartDashboardClient() error {
	if messageBrokerClient.dashboardClient == nil {
		return Error.New("Dashboard client is not configured", nil)
	}
	return messageBrokerClient.dashboardClient.Start()
}

func (messageBrokerClient *Client) StopDashboardClient() error {
	if messageBrokerClient.dashboardClient == nil {
		return Error.New("Dashboard client is not configured", nil)
	}
	return messageBrokerClient.dashboardClient.Stop()
}

func (messageBrokerClient *Client) Start() error {
	messageBrokerClient.statusMutex.Lock()
	defer messageBrokerClient.statusMutex.Unlock()
	if messageBrokerClient.status != Status.STOPPED {
		return Error.New("Already started", nil)
	}
	messageBrokerClient.status = Status.PENDING
	stopChannel := make(chan bool)
	messageBrokerClient.stopChannel = stopChannel

	for topic := range messageBrokerClient.subscribedAsyncTopics {
		resolutionAttempt, _ := messageBrokerClient.startResolutionAttempt(topic, false, stopChannel, messageBrokerClient.subscribedAsyncTopics[topic])
		<-resolutionAttempt.ongoing
	}
	for topic := range messageBrokerClient.subscribedSyncTopics {
		resolutionAttempt, _ := messageBrokerClient.startResolutionAttempt(topic, true, stopChannel, messageBrokerClient.subscribedSyncTopics[topic])
		<-resolutionAttempt.ongoing
	}
	messageBrokerClient.status = Status.STARTED
	return nil
}

func (messageBrokerClient *Client) stop() {
	close(messageBrokerClient.stopChannel)
	messageBrokerClient.stopChannel = nil
	messageBrokerClient.waitGroup.Wait()
	messageBrokerClient.status = Status.STOPPED
}
func (messageBrokerClient *Client) Stop() error {
	messageBrokerClient.statusMutex.Lock()
	defer messageBrokerClient.statusMutex.Unlock()
	if messageBrokerClient.status != Status.STARTED {
		return Error.New("Already started", nil)
	}
	messageBrokerClient.status = Status.PENDING
	messageBrokerClient.stop()
	return nil
}

func (messageBrokerClient *Client) GetStatus() int {
	return messageBrokerClient.status
}

func (messageBrokerClient *Client) GetMetrics() map[string]uint64 {
	m := map[string]uint64{}
	messageBrokerClient.mutex.Lock()
	m["ongoingTopicResolutions"] = uint64(len(messageBrokerClient.ongoingTopicResolutions))
	m["brokerConnections"] = uint64(len(messageBrokerClient.brokerConnections))
	m["topicResolutions"] = uint64(len(messageBrokerClient.topicResolutions))
	for _, connection := range messageBrokerClient.brokerConnections {
		metrics := connection.connection.RetrieveMetrics()
		for key, value := range metrics {
			println(key, value)
			m[key] += value
		}
	}
	messageBrokerClient.mutex.Unlock()
	return m
}

func (messageBrokerClient *Client) GetName() string {
	return messageBrokerClient.name
}

func getEndpointString(endpoint *Config.TcpClient) string {
	return endpoint.Address + endpoint.TlsCert
}

func getASyncString(async bool) string {
	if async {
		return "async"
	}
	return "sync"
}
