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

	brokerConnections map[string]*connection // endpointString -> connection
	topicResolutions  map[string]*connection // topic -> connection

	mutex sync.Mutex

	asyncTopics map[string]bool
	syncTopics  map[string]bool
}

type connection struct {
	connection *SystemgeConnection.SystemgeConnection
	endpoint   *Config.TcpEndpoint
	topics     map[string]bool
}

type resultionAttempt struct {
	ongoing chan bool
	result  *connection
}

func NewMessageBrokerClient_(config *Config.MessageBrokerClient, systemgeMessageHandler SystemgeConnection.MessageHandler, dashboardCommands Commands.Handlers) *MessageBrokerClient {
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
	if config.ResolverConnectionConfig.TcpBufferBytes == 0 {
		config.ResolverConnectionConfig.TcpBufferBytes = 1024 * 4
	}
	if config.ConnectionConfig.TcpBufferBytes == 0 {
		config.ConnectionConfig.TcpBufferBytes = 1024 * 4
	}

	messageBrokerClient := &MessageBrokerClient{
		config:                  config,
		messageHandler:          systemgeMessageHandler,
		ongoingTopicResolutions: make(map[string]*resultionAttempt),

		topicResolutions: make(map[string]*connection),

		brokerConnections: make(map[string]*connection),

		asyncTopics: make(map[string]bool),
		syncTopics:  make(map[string]bool),

		status: Status.STOPPED,
	}
	if config.InfoLoggerPath != "" {
		messageBrokerClient.infoLogger = Tools.NewLogger("[Info: \""+config.Name+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		messageBrokerClient.warningLogger = Tools.NewLogger("[Warning: \""+config.Name+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		messageBrokerClient.errorLogger = Tools.NewLogger("[Error: \""+config.Name+"\"] ", config.ErrorLoggerPath)
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
		connection, err := messageBrokerClient.resolveConnection(topic)
	}
	for topic, _ := range messageBrokerClient.syncTopics {
		connection, err := messageBrokerClient.resolveConnection(topic)
	}

	messageBrokerClient.status = Status.STARTED
	return nil
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

// TODO (different func): on disconnect or if lifetime >0, wait for lifetime to pass and resolve topic again. if endpoint changes, update connection and subscribe to topic.

// subscribes to the provided topic using the provided connection
func (MessageBrokerClient *MessageBrokerClient) subscribeToTopic(connection *connection, topic string, sync bool) error {
	if _, exists := connection.topics[topic]; exists {
		// should this ever be called if topic is already subscribed?
		return nil // possibly return error
	}
	if sync {
		_, err := connection.connection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_SYNC, topic)
		if err != nil {
			// todo: clean up connection if no other topics (issues due to concurrency?)
			return Error.New("Failed to subscribe to topic", err)
		}
	} else {
		_, err := connection.connection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_ASYNC, topic)
		if err != nil {
			// todo: clean up connection if no other topics (issues due to concurrency?)
			return Error.New("Failed to subscribe to topic", err)
		}
	}
	connection.topics[topic] = true
	return nil
}

// goes through all assigned resolvers in order and tries to resolve the topic.
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

// checks if connection to endpoint is already established. if not, establishes connection.
func (messageBrokerClient *MessageBrokerClient) getConnection(endpoint *Config.TcpEndpoint) (*connection, error) {
	messageBrokerClient.mutex.Lock()
	conn := messageBrokerClient.brokerConnections[getEndpointString(endpoint)]
	messageBrokerClient.mutex.Unlock()

	if conn == nil {
		newConnection, err := SystemgeConnection.EstablishConnection(messageBrokerClient.config.ConnectionConfig, endpoint, messageBrokerClient.GetName(), messageBrokerClient.config.MaxServerNameLength)
		if err != nil {
			return nil, Error.New("Failed to establish connection to broker", err)
		}
		conn = &connection{
			connection: newConnection,
			endpoint:   endpoint,
			topics:     map[string]bool{},
		}
		messageBrokerClient.mutex.Lock()
		messageBrokerClient.brokerConnections[getEndpointString(endpoint)] = conn
		messageBrokerClient.mutex.Unlock()
	}
	return conn, nil
}

// checks if connection for topic is already resolved. if not, checks if resolution is ongoing and waits for it to finish.
// if resolution is not ongoing, starts resolution by resolving the topics endpoint.
// checks if connection to endpoint is already established. if not, establishes connection.
func (messageBrokerClient *MessageBrokerClient) resolveConnection(topic string) (*connection, error) {
	messageBrokerClient.mutex.Lock()
	if resolution := messageBrokerClient.topicResolutions[topic]; resolution != nil {
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

	finishAttempt := func(result *connection) {
		messageBrokerClient.mutex.Lock()
		resolutionAttempt.result = result
		close(resolutionAttempt.ongoing)
		delete(messageBrokerClient.ongoingTopicResolutions, topic)
		if result != nil {
			messageBrokerClient.topicResolutions[topic] = result
			result.topics[topic] = true
			/* go func() {
				var topicResolutionTimeout <-chan time.Time
				if messageBrokerClient.config.TopicResolutionLifetimeMs > 0 {
					topicResolutionTimeout = time.After(time.Duration(messageBrokerClient.config.TopicResolutionLifetimeMs) * time.Millisecond)
				}
				select {
				case <-topicResolutionTimeout:

				case <-connection.connection.GetCloseChannel():

				}
			}() */
		}
		messageBrokerClient.mutex.Unlock()
	}

	brokerEndpoint, err := messageBrokerClient.resolveBrokerEndpoint(topic)
	if err != nil {
		finishAttempt(nil)
		return nil, Error.New("Failed to resolve broker endpoint", err)
	}

	connection, err := messageBrokerClient.getConnection(brokerEndpoint)
	if err != nil {
		finishAttempt(nil)
		return nil, Error.New("Failed to get connection", err)
	}

	finishAttempt(connection)
	return resolutionAttempt.result, nil
}

func getEndpointString(endpoint *Config.TcpEndpoint) string {
	return endpoint.Address + endpoint.TlsCert
}
