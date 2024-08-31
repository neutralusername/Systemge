package MessageBroker

import (
	"sync"
	"time"

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
		_, err := messageBrokerClient.resolveConnection(topic, false)

	}
	for topic, _ := range messageBrokerClient.syncTopics {
		_, err := messageBrokerClient.resolveConnection(topic, true)

	}

	messageBrokerClient.status = Status.STARTED
	return nil
}

func (messageBrokerClient *MessageBrokerClient) Stop() error {
	messageBrokerClient.statusMutex.Lock()
	defer messageBrokerClient.statusMutex.Unlock()
	if messageBrokerClient.status != Status.STARTED {
		return Error.New("Already started", nil)
	}
	messageBrokerClient.status = Status.PENDING

	// make sure no new connection attempts can be initiated or completed and all ongoing attempts are finished from this point onwards.

	messageBrokerClient.mutex.Lock()
	for _, connection := range messageBrokerClient.brokerConnections {
		connection.connection.Close()
	}
	messageBrokerClient.mutex.Unlock()

	messageBrokerClient.status = Status.STOPPED
	return nil
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
		resolverConnection.Close() // close in case there was an issue on the resolver side that prevented closing the connection
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

func (messageBrokerClient *MessageBrokerClient) resolveConnection(topic string, syncTopic bool) (*connection, error) {
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
		if result != nil {
			messageBrokerClient.topicResolutions[topic] = result
			result.topics[topic] = true
			messageBrokerClient.brokerConnections[getEndpointString(result.endpoint)] = result // operation can be redundant if connection was already established for another topic
			subscribedTopic := (syncTopic && messageBrokerClient.syncTopics[topic]) || (!syncTopic && messageBrokerClient.asyncTopics[topic])
			if subscribedTopic {
				if err := messageBrokerClient.subscribeToTopic(result, topic, syncTopic); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Error.New("Failed to subscribe to "+getASyncString(syncTopic)+" topic \""+topic+"\" on broker \""+result.endpoint.Address+"\"", err).Error())
					}
					if messageBrokerClient.mailer != nil {
						if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to subscribe to "+getASyncString(syncTopic)+" topic \""+topic+"\" on broker \""+result.endpoint.Address+"\"", err).Error())); err != nil {
							if messageBrokerClient.errorLogger != nil {
								messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
							}
						}
					}
				}
			}
			go messageBrokerClient.handleTopicResolutionLifetime(result, topic, subscribedTopic)
		}
		delete(messageBrokerClient.ongoingTopicResolutions, topic)
		resolutionAttempt.result = result
		close(resolutionAttempt.ongoing)
		messageBrokerClient.mutex.Unlock()
	}

	endpoint, err := messageBrokerClient.resolveBrokerEndpoint(topic)
	if err != nil {
		finishAttempt(nil)
		return nil, Error.New("Failed to resolve broker endpoint", err)
	}

	messageBrokerClient.mutex.Lock()
	existingConnection := messageBrokerClient.brokerConnections[getEndpointString(endpoint)]
	messageBrokerClient.mutex.Unlock()

	if existingConnection != nil {
		finishAttempt(existingConnection)
		return existingConnection, nil
	}

	systemgeConnection, err := SystemgeConnection.EstablishConnection(messageBrokerClient.config.ConnectionConfig, endpoint, messageBrokerClient.GetName(), messageBrokerClient.config.MaxServerNameLength)
	if err != nil {
		finishAttempt(nil)
		return nil, Error.New("Failed to establish connection to broker", err)
	}

	finishAttempt(&connection{
		connection: systemgeConnection,
		endpoint:   endpoint,
		topics:     map[string]bool{},
	})
	return resolutionAttempt.result, nil
}

func (messageBrokerClient *MessageBrokerClient) handleTopicResolutionLifetime(connection *connection, topic string, subscribedTopic bool) {
	var topicResolutionTimeout <-chan time.Time
	if messageBrokerClient.config.TopicResolutionLifetimeMs > 0 {
		topicResolutionTimeout = time.After(time.Duration(messageBrokerClient.config.TopicResolutionLifetimeMs) * time.Millisecond)
	}
	select {
	case <-topicResolutionTimeout:
		if subscribedTopic {

		} else {

		}
	case <-connection.connection.GetCloseChannel():
		messageBrokerClient.mutex.Lock()
		subscribedAsyncTopicsByClosedConnection := []string{}
		subscribedSyncTopicsByClosedConnection := []string{}
		for topic, _ := range connection.topics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.topics, topic)
			if messageBrokerClient.asyncTopics[topic] {
				subscribedAsyncTopicsByClosedConnection = append(subscribedAsyncTopicsByClosedConnection, topic)
			} else if messageBrokerClient.syncTopics[topic] {
				subscribedSyncTopicsByClosedConnection = append(subscribedSyncTopicsByClosedConnection, topic)
			}
		}
		delete(messageBrokerClient.brokerConnections, getEndpointString(connection.endpoint))
		messageBrokerClient.mutex.Unlock()

		messageBrokerClient.mutex.Lock()
		defer messageBrokerClient.mutex.Unlock()
		if messageBrokerClient.status != Status.STARTED {
			return
		}
		for _, topic := range subscribedAsyncTopicsByClosedConnection {

			_, err := messageBrokerClient.resolveConnection(topic, false)

		}
		for _, topic := range subscribedSyncTopicsByClosedConnection {

			_, err := messageBrokerClient.resolveConnection(topic, true)

		}
	}
}

func (MessageBrokerClient *MessageBrokerClient) subscribeToTopic(connection *connection, topic string, sync bool) error {
	if sync {
		_, err := connection.connection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_SYNC, topic)
		if err != nil {
			return Error.New("Failed to subscribe to topic", err)
		}
	} else {
		_, err := connection.connection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_ASYNC, topic)
		if err != nil {
			return Error.New("Failed to subscribe to topic", err)
		}
	}
	return nil
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
