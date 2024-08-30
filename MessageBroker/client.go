package MessageBroker

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeClient"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type MessageBrokerClient struct {
	status      int
	statusMutex sync.Mutex

	config *Config.MessageBrokerClient

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	messageHandler SystemgeConnection.MessageHandler

	messageBrokerClient *SystemgeClient.SystemgeClient
	dashboardClient     *Dashboard.DashboardClient

	resolverConnections map[string]*SystemgeConnection.SystemgeConnection // topic -> connection

	topicResolutions map[string]map[string]*SystemgeConnection.SystemgeConnection // topic -> [connectionName -> connection]

	asyncTopics map[string]bool
	syncTopics  map[string]bool
	mutex       sync.Mutex
}

func NewMessageBrokerClient_(config *Config.MessageBrokerClient, systemgeMessageHandler SystemgeConnection.MessageHandler, dashboardCommands Commands.Handlers) *MessageBrokerClient {
	if config == nil {
		panic(Error.New("Config is required", nil))
	}
	if config.ResolverConnectionConfig == nil {
		panic(Error.New("ResolverConnectionConfig is required", nil))
	}
	if config.ResolverEndpoint == nil {
		panic(Error.New("ResolverEndpoint is required", nil))
	}
	if config.MessageBrokerClientConfig == nil {
		panic(Error.New("MessageBrokerClientConfig is required", nil))
	}
	if config.MessageBrokerClientConfig.ConnectionConfig == nil {
		panic(Error.New("MessageBrokerClientConfig.ConnectionConfig is required", nil))
	}
	if config.MessageBrokerClientConfig.Name == "" {
		panic(Error.New("MessageBrokerClientConfig.Name is required", nil))
	}
	if config.MessageBrokerClientConfig.ConnectionConfig.TcpBufferBytes == 0 {
		config.MessageBrokerClientConfig.ConnectionConfig.TcpBufferBytes = 1024 * 4
	}
	if config.ResolverConnectionConfig.TcpBufferBytes == 0 {
		config.ResolverConnectionConfig.TcpBufferBytes = 1024 * 4
	}

	messageBrokerClient := &MessageBrokerClient{
		config:           config,
		messageHandler:   systemgeMessageHandler,
		topicResolutions: make(map[string]map[string]*SystemgeConnection.SystemgeConnection),

		asyncTopics: make(map[string]bool),
		syncTopics:  make(map[string]bool),

		status: Status.STOPPED,
	}
	if config.InfoLoggerPath != "" {
		messageBrokerClient.infoLogger = Tools.NewLogger("[Info: \"MessageBrokerClient\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		messageBrokerClient.warningLogger = Tools.NewLogger("[Warning: \"MessageBrokerClient\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		messageBrokerClient.errorLogger = Tools.NewLogger("[Error: \"MessageBrokerClient\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		messageBrokerClient.mailer = Tools.NewMailer(config.MailerConfig)
	}

	messageBrokerClient.messageBrokerClient = SystemgeClient.New(config.MessageBrokerClientConfig, nil, nil)

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
		messageBrokerClient.asyncTopics[asyncTopic] = true
	}
	for _, syncTopic := range config.SyncTopics {
		messageBrokerClient.syncTopics[syncTopic] = true
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

}

func (messageBrokerClient *MessageBrokerClient) Stop() error {

}

func (messageBrokerClient *MessageBrokerClient) GetStatus() int {

}

func (messageBrokerClient *MessageBrokerClient) GetMetrics() map[string]uint64 {}
