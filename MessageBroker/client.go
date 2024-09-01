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

	waitGroup sync.WaitGroup

	messageHandler SystemgeConnection.MessageHandler

	dashboardClient *Dashboard.DashboardClient

	ongoingTopicResolutions map[string]*resolutionAttempt

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

type resolutionAttempt struct {
	topic         string
	syncTopic     bool
	ongoing       chan bool
	result        *connection
	newConnection bool
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
		ongoingTopicResolutions: make(map[string]*resolutionAttempt),

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

	go func() {
		for topic, _ := range messageBrokerClient.asyncTopics {
			err := messageBrokerClient.startResolutionAttempt(topic, false)

		}
		for topic, _ := range messageBrokerClient.syncTopics {
			err := messageBrokerClient.startResolutionAttempt(topic, true)

		}
	}()

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

	messageBrokerClient.waitGroup.Wait()

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

func (messageBrokerClient *MessageBrokerClient) startResolutionAttempt(topic string, syncTopic bool) error {
	messageBrokerClient.statusMutex.Lock()
	if messageBrokerClient.status == Status.STOPPED {
		messageBrokerClient.statusMutex.Unlock()
		return Error.New("Client is stopped", nil)
	}

	messageBrokerClient.waitGroup.Add(1)

	messageBrokerClient.mutex.Lock()
	if resolution := messageBrokerClient.topicResolutions[topic]; resolution != nil {
		messageBrokerClient.mutex.Unlock()
		messageBrokerClient.statusMutex.Unlock()

		return nil
	}
	if resolutionAttempt := messageBrokerClient.ongoingTopicResolutions[topic]; resolutionAttempt != nil {
		messageBrokerClient.mutex.Unlock()
		messageBrokerClient.statusMutex.Unlock()

		<-resolutionAttempt.ongoing
		if resolutionAttempt.result == nil {
			return Error.New("Failed to resolve connection", nil)
		}
		return nil
	}
	resolutionAttempt := &resolutionAttempt{
		ongoing:   make(chan bool),
		topic:     topic,
		syncTopic: syncTopic,
	}
	messageBrokerClient.ongoingTopicResolutions[topic] = resolutionAttempt
	messageBrokerClient.mutex.Unlock()
	messageBrokerClient.statusMutex.Unlock()

	go func() {
		err := messageBrokerClient.resolutionAttempt(resolutionAttempt)
	}()
	return nil
}

func (messageBrokerClient *MessageBrokerClient) resolutionAttempt(resolutionAttempt *resolutionAttempt) error {
	defer messageBrokerClient.finishResolutionAttempt(resolutionAttempt)

	endpoint, err := messageBrokerClient.resolveBrokerEndpoint(resolutionAttempt.topic)
	if err != nil {
		return Error.New("Failed to resolve broker endpoint", err)
	}

	messageBrokerClient.mutex.Lock()
	existingConnection := messageBrokerClient.brokerConnections[getEndpointString(endpoint)]
	messageBrokerClient.mutex.Unlock()

	if existingConnection != nil {
		resolutionAttempt.result = existingConnection
		return nil
	}
	systemgeConnection, err := SystemgeConnection.EstablishConnection(messageBrokerClient.config.ConnectionConfig, endpoint, messageBrokerClient.GetName(), messageBrokerClient.config.MaxServerNameLength)
	if err != nil {
		return Error.New("Failed to establish connection to broker", err)
	}
	connection := &connection{
		connection: systemgeConnection,
		endpoint:   endpoint,
		topics:     map[string]bool{},
	}
	resolutionAttempt.result = connection
	resolutionAttempt.newConnection = true
	return nil
}

func (messageBrokerClient *MessageBrokerClient) finishResolutionAttempt(resolutionAttempt *resolutionAttempt) {
	if resolutionAttempt.result == nil {
		if (resolutionAttempt.syncTopic && messageBrokerClient.syncTopics[resolutionAttempt.topic]) || (!resolutionAttempt.syncTopic && messageBrokerClient.asyncTopics[resolutionAttempt.topic]) {
			// will let other goroutines waiting until the resolution attempt for this topic is finished as of now
			go messageBrokerClient.resolutionAttempt(resolutionAttempt)
			// todo: make sure this doesn't result in infinite loop in case of messageBrokerClient.Stop() call
		} else {
			messageBrokerClient.mutex.Lock()
			delete(messageBrokerClient.ongoingTopicResolutions, resolutionAttempt.topic)
			messageBrokerClient.mutex.Unlock()
			close(resolutionAttempt.ongoing)
			messageBrokerClient.waitGroup.Done()
		}
	} else {
		messageBrokerClient.mutex.Lock()
		messageBrokerClient.topicResolutions[resolutionAttempt.topic] = resolutionAttempt.result
		resolutionAttempt.result.topics[resolutionAttempt.topic] = true
		if resolutionAttempt.newConnection {
			messageBrokerClient.brokerConnections[getEndpointString(resolutionAttempt.result.endpoint)] = resolutionAttempt.result
			go messageBrokerClient.handleConnectionLifetime(resolutionAttempt.result)
		}
		go messageBrokerClient.handleTopicResolutionLifetime(resolutionAttempt.result, resolutionAttempt.topic, resolutionAttempt.syncTopic)
		delete(messageBrokerClient.ongoingTopicResolutions, resolutionAttempt.topic)
		messageBrokerClient.mutex.Unlock()

		if (resolutionAttempt.syncTopic && messageBrokerClient.syncTopics[resolutionAttempt.topic]) || (!resolutionAttempt.syncTopic && messageBrokerClient.asyncTopics[resolutionAttempt.topic]) {
			if err := messageBrokerClient.subscribeToTopic(resolutionAttempt.result, resolutionAttempt.topic, resolutionAttempt.syncTopic); err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Error.New("Failed to subscribe to "+getASyncString(resolutionAttempt.syncTopic)+" topic \""+resolutionAttempt.topic+"\" on broker \""+resolutionAttempt.result.endpoint.Address+"\"", err).Error())
				}
				if messageBrokerClient.mailer != nil {
					if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to subscribe to "+getASyncString(resolutionAttempt.syncTopic)+" topic \""+resolutionAttempt.topic+"\" on broker \""+resolutionAttempt.result.endpoint.Address+"\"", err).Error())); err != nil {
						if messageBrokerClient.errorLogger != nil {
							messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
						}
					}
				}
			}
		}
		close(resolutionAttempt.ongoing)
		messageBrokerClient.waitGroup.Done()
	}
}

func (messageBrokerClient *MessageBrokerClient) handleTopicResolutionLifetime(connection *connection, topic string, syncTopic bool) {
	var topicResolutionTimeout <-chan time.Time
	if messageBrokerClient.config.TopicResolutionLifetimeMs > 0 {
		topicResolutionTimeout = time.After(time.Duration(messageBrokerClient.config.TopicResolutionLifetimeMs) * time.Millisecond)
	}
	select {
	case <-topicResolutionTimeout:
		messageBrokerClient.mutex.Lock()
		delete(messageBrokerClient.topicResolutions, topic)
		delete(connection.topics, topic)

		if len(connection.topics) == 0 {
			delete(messageBrokerClient.brokerConnections, getEndpointString(connection.endpoint))
			connection.connection.Close()
		}
		messageBrokerClient.mutex.Unlock()

		// todo: finish this

		if (syncTopic && messageBrokerClient.syncTopics[topic]) || (!syncTopic && messageBrokerClient.asyncTopics[topic]) {

		} else {

		}
	}
}

func (messageBrokerClient *MessageBrokerClient) handleConnectionLifetime(connection *connection) {
	select {
	case <-connection.connection.GetCloseChannel():
		// todo: think through possible race conditions between the go call and the select and in regards to topic lifetime handling

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

		for _, topic := range subscribedAsyncTopicsByClosedConnection {
			err := messageBrokerClient.startResolutionAttempt(topic, false)

		}
		for _, topic := range subscribedSyncTopicsByClosedConnection {
			err := messageBrokerClient.startResolutionAttempt(topic, true)

		}
	}
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
