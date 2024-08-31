package MessageBroker

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
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

	dashboardClient *Dashboard.DashboardClient

	ongoingTopicResolutions map[string]*resultionAttempt
	outTopicResolution      map[string]*SystemgeConnection.SystemgeConnection // topic -> connection
	inTopicResolution       map[string]*SystemgeConnection.SystemgeConnection // topic -> connection

	mutex sync.Mutex

	asyncTopics map[string]bool
	syncTopics  map[string]bool
}

type resultionAttempt struct {
	ongoing chan bool
	result  *SystemgeConnection.SystemgeConnection
}

func NewMessageBrokerClient_(config *Config.MessageBrokerClient, systemgeMessageHandler SystemgeConnection.MessageHandler, dashboardCommands Commands.Handlers) *MessageBrokerClient {
	if config == nil {
		panic(Error.New("Config is required", nil))
	}
	if config.ResolverConnectionConfig == nil {
		panic(Error.New("ResolverConnectionConfig is required", nil))
	}
	if len(config.ResolverEndpoints) == 0 {
		panic(Error.New("At least one ResolverEndpoint is required", nil))
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
		config:             config,
		messageHandler:     systemgeMessageHandler,
		outTopicResolution: make(map[string]*SystemgeConnection.SystemgeConnection),

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
	messageBrokerClient.statusMutex.Lock()
	defer messageBrokerClient.statusMutex.Unlock()
	if messageBrokerClient.status != Status.STOPPED {
		return Error.New("Already started", nil)
	}
	messageBrokerClient.status = Status.PENDING

	for topic, _ := range messageBrokerClient.asyncTopics {
		endpoint, err := messageBrokerClient.resolveBrokerEndpoint(topic)
		if err != nil {
			messageBrokerClient.status = Status.STOPPED
			return Error.New("Failed to resolve broker endpoint", err)
		}

	}
	for topic, _ := range messageBrokerClient.syncTopics {
		endpoint, err := messageBrokerClient.resolveBrokerEndpoint(topic)
		if err != nil {
			messageBrokerClient.status = Status.STOPPED
			return Error.New("Failed to resolve broker endpoint", err)
		}

	}

	messageBrokerClient.status = Status.STARTED
	return nil
}

// checks if connection to endpoint exists. if not exists establish connection. use connection to subscribe to topic.
// if lifetime >0, wait for lifetime to pass and resolve topic again. if endpoint changes, update connection and subscribe to topic.
func (messageBrokerClient *MessageBrokerClient) subscriptionLoop(endpoint *Config.TcpEndpoint, topic string) {

}

func (messageBrokerClient *MessageBrokerClient) Stop() error {

}

func (messageBrokerClient *MessageBrokerClient) GetStatus() int {

}

func (messageBrokerClient *MessageBrokerClient) GetMetrics() map[string]uint64 {

}

func (messageBrokerClient *MessageBrokerClient) GetName() string {
	return messageBrokerClient.config.Name
}

func (messageBrokerclient *MessageBrokerClient) resolveBrokerEndpoint(topic string) (*Config.TcpEndpoint, error) {
	for _, resolverEndpoint := range messageBrokerclient.config.ResolverEndpoints {
		resolverConnection, err := SystemgeConnection.EstablishConnection(messageBrokerclient.config.ResolverConnectionConfig, resolverEndpoint, messageBrokerclient.GetName(), messageBrokerclient.config.MaxServerNameLength)
		if err != nil {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Error.New("Failed to establish connection to resolver \""+resolverEndpoint.Address+"\"", err).Error())
			}
			continue
		}
		response, err := resolverConnection.SyncRequestBlocking(Message.TOPIC_RESOLVE_ASYNC, topic)
		resolverConnection.Close() //probably redundant
		if err != nil {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Error.New("Failed to send resolution request to resolver \""+resolverEndpoint.Address+"\"", err).Error())
			}
			continue
		}
		if response.GetTopic() == Message.TOPIC_FAILURE {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Error.New("Failed to resolve topic \""+topic+"\" using resolver \""+resolverEndpoint.Address+"\"", nil).Error())
			}
			continue
		}
		endpoint := Config.UnmarshalTcpEndpoint(response.GetPayload())
		if endpoint == nil {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Error.New("Failed to unmarshal endpoint", nil).Error())
			}
			continue
		}
		return endpoint, nil
	}
	return nil, Error.New("Failed to resolve broker endpoint", nil)
}

func (messageBrokerClient *MessageBrokerClient) resolveConnection(topic string) (*SystemgeConnection.SystemgeConnection, error) {
	messageBrokerClient.mutex.Lock()
	if resolution := messageBrokerClient.outTopicResolution[topic]; resolution != nil {
		messageBrokerClient.mutex.Unlock()
		return resolution, nil
	}
	if resolutionAttempt := messageBrokerClient.ongoingTopicResolutions[topic]; resolutionAttempt != nil {
		messageBrokerClient.mutex.Unlock()
		<-resolutionAttempt.ongoing
		if resolutionAttempt.result == nil {
			return nil, Error.New("Failed to resolve connection", nil)
		}
		return resolutionAttempt.result, nil
	}
	resolutionAttempt := &resultionAttempt{
		ongoing: make(chan bool),
	}
	messageBrokerClient.ongoingTopicResolutions[topic] = resolutionAttempt
	messageBrokerClient.mutex.Unlock()

	brokerEndpoint, err := messageBrokerClient.resolveBrokerEndpoint(topic)
	if err != nil {
		messageBrokerClient.mutex.Lock()
		close(resolutionAttempt.ongoing)
		delete(messageBrokerClient.ongoingTopicResolutions, topic)
		messageBrokerClient.mutex.Unlock()
		return nil, Error.New("Failed to resolve broker endpoint", err)
	}
	brokerConnection, err := SystemgeConnection.EstablishConnection(messageBrokerClient.config.MessageBrokerClientConfig.ConnectionConfig, brokerEndpoint, messageBrokerClient.GetName(), messageBrokerClient.config.MaxServerNameLength)
	if err != nil {
		messageBrokerClient.mutex.Lock()
		close(resolutionAttempt.ongoing)
		delete(messageBrokerClient.ongoingTopicResolutions, topic)
		messageBrokerClient.mutex.Unlock()
		return nil, Error.New("Failed to establish connection to broker", err)
	}
	messageBrokerClient.mutex.Lock()
	resolutionAttempt.result = brokerConnection
	close(resolutionAttempt.ongoing)
	delete(messageBrokerClient.ongoingTopicResolutions, topic)
	messageBrokerClient.mutex.Unlock()
	return brokerConnection, nil
}
