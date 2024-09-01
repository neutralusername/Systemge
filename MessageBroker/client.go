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

	stopChannel chan bool

	messageHandler SystemgeConnection.MessageHandler

	dashboardClient *Dashboard.DashboardClient

	ongoingTopicResolutions map[string]*resolutionAttempt

	brokerConnections map[string]*connection // endpointString -> connection
	topicResolutions  map[string]*connection // topic -> connection

	mutex sync.Mutex

	subscribedAsyncTopics map[string]bool
	subscribedSyncTopics  map[string]bool
}

type connection struct {
	connection             *SystemgeConnection.SystemgeConnection
	endpoint               *Config.TcpEndpoint
	responsibleAsyncTopics map[string]bool
	responsibleSyncTopics  map[string]bool
}

type resolutionAttempt struct {
	topic         string
	isSyncTopic   bool
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

		subscribedAsyncTopics: make(map[string]bool),
		subscribedSyncTopics:  make(map[string]bool),

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
		messageBrokerClient.subscribedAsyncTopics[asyncTopic] = true
	}
	for _, syncTopic := range config.SyncTopics {
		messageBrokerClient.subscribedSyncTopics[syncTopic] = true
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

	go func() {
		for topic, _ := range messageBrokerClient.subscribedAsyncTopics {
			err := messageBrokerClient.startResolutionAttempt(topic, false, stopChannel)

		}
		for topic, _ := range messageBrokerClient.subscribedSyncTopics {
			err := messageBrokerClient.startResolutionAttempt(topic, true, stopChannel)

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

	close(messageBrokerClient.stopChannel)
	messageBrokerClient.stopChannel = nil
	messageBrokerClient.waitGroup.Wait()

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

func (messageBrokerClient *MessageBrokerClient) startResolutionAttempt(topic string, syncTopic bool, stopChannel chan bool) error {
	messageBrokerClient.statusMutex.Lock()
	if stopChannel != messageBrokerClient.stopChannel {
		messageBrokerClient.statusMutex.Unlock()
		return Error.New("Aborted because resolution attempt is from previous session", nil)
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
		ongoing:     make(chan bool),
		topic:       topic,
		isSyncTopic: syncTopic,
	}
	messageBrokerClient.ongoingTopicResolutions[topic] = resolutionAttempt
	messageBrokerClient.mutex.Unlock()
	messageBrokerClient.statusMutex.Unlock()

	go func() {
		err := messageBrokerClient.resolutionAttempt(resolutionAttempt, stopChannel)
	}()
	return nil
}

func (messageBrokerClient *MessageBrokerClient) resolutionAttempt(resolutionAttempt *resolutionAttempt, stopChannel chan bool) error {
	defer messageBrokerClient.finishResolutionAttempt(resolutionAttempt, stopChannel)

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
		connection:             systemgeConnection,
		endpoint:               endpoint,
		responsibleAsyncTopics: make(map[string]bool),
		responsibleSyncTopics:  make(map[string]bool),
	}
	resolutionAttempt.result = connection
	resolutionAttempt.newConnection = true
	return nil
}

func (messageBrokerClient *MessageBrokerClient) finishResolutionAttempt(resolutionAttempt *resolutionAttempt, stopChannel chan bool) {
	if resolutionAttempt.result == nil {
		if (resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedSyncTopics[resolutionAttempt.topic]) || (!resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedAsyncTopics[resolutionAttempt.topic]) {
			// will let other goroutines waiting until the resolution attempt for this topic is finished as of now
			// could result in a suboptimal situation when the client is trying to message the broker responsible for this topic
			select {
			default:
				go func() {
					err := messageBrokerClient.resolutionAttempt(resolutionAttempt, stopChannel)
				}()
				return
			case <-stopChannel:
			}
		}
		messageBrokerClient.mutex.Lock()
		delete(messageBrokerClient.ongoingTopicResolutions, resolutionAttempt.topic)
		messageBrokerClient.mutex.Unlock()
		close(resolutionAttempt.ongoing)
		messageBrokerClient.waitGroup.Done()
	} else {
		messageBrokerClient.mutex.Lock()
		messageBrokerClient.topicResolutions[resolutionAttempt.topic] = resolutionAttempt.result
		if resolutionAttempt.isSyncTopic {
			resolutionAttempt.result.responsibleSyncTopics[resolutionAttempt.topic] = true
		} else {
			resolutionAttempt.result.responsibleAsyncTopics[resolutionAttempt.topic] = true
		}
		if resolutionAttempt.newConnection {
			messageBrokerClient.brokerConnections[getEndpointString(resolutionAttempt.result.endpoint)] = resolutionAttempt.result
			go messageBrokerClient.handleConnectionLifetime(resolutionAttempt.result, stopChannel)
		}
		go messageBrokerClient.handleTopicResolutionLifetime(resolutionAttempt.result, resolutionAttempt.topic, resolutionAttempt.isSyncTopic, stopChannel)
		delete(messageBrokerClient.ongoingTopicResolutions, resolutionAttempt.topic)
		messageBrokerClient.mutex.Unlock()

		if (resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedSyncTopics[resolutionAttempt.topic]) || (!resolutionAttempt.isSyncTopic && messageBrokerClient.subscribedAsyncTopics[resolutionAttempt.topic]) {
			if err := messageBrokerClient.subscribeToTopic(resolutionAttempt.result, resolutionAttempt.topic, resolutionAttempt.isSyncTopic); err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Error.New("Failed to subscribe to "+getASyncString(resolutionAttempt.isSyncTopic)+" topic \""+resolutionAttempt.topic+"\" on broker \""+resolutionAttempt.result.endpoint.Address+"\"", err).Error())
				}
				if messageBrokerClient.mailer != nil {
					if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to subscribe to "+getASyncString(resolutionAttempt.isSyncTopic)+" topic \""+resolutionAttempt.topic+"\" on broker \""+resolutionAttempt.result.endpoint.Address+"\"", err).Error())); err != nil {
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

func (messageBrokerClient *MessageBrokerClient) handleTopicResolutionLifetime(connection *connection, topic string, syncTopic bool, stopChannel chan bool) {
	var topicResolutionTimeout <-chan time.Time
	if messageBrokerClient.config.TopicResolutionLifetimeMs > 0 {
		topicResolutionTimeout = time.After(time.Duration(messageBrokerClient.config.TopicResolutionLifetimeMs) * time.Millisecond)
	}
	select {
	case <-topicResolutionTimeout:
		messageBrokerClient.mutex.Lock()
		delete(messageBrokerClient.topicResolutions, topic)
		if syncTopic {
			delete(connection.responsibleSyncTopics, topic)
		} else {
			delete(connection.responsibleAsyncTopics, topic)
		}
		if len(connection.responsibleAsyncTopics) == 0 && len(connection.responsibleSyncTopics) == 0 {
			delete(messageBrokerClient.brokerConnections, getEndpointString(connection.endpoint))
			connection.connection.Close()
		}
		messageBrokerClient.mutex.Unlock()
		if (syncTopic && messageBrokerClient.subscribedSyncTopics[topic]) || (!syncTopic && messageBrokerClient.subscribedAsyncTopics[topic]) {
			err := messageBrokerClient.startResolutionAttempt(topic, syncTopic, stopChannel)
		}
	case <-stopChannel:
		return
	}
}

func (messageBrokerClient *MessageBrokerClient) handleConnectionLifetime(connection *connection, stopChannel chan bool) {
	select {
	case <-connection.connection.GetCloseChannel():
		messageBrokerClient.mutex.Lock()
		subscribedAsyncTopicsByClosedConnection := []string{}
		subscribedSyncTopicsByClosedConnection := []string{}
		for topic, _ := range connection.responsibleAsyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleAsyncTopics, topic)
			if messageBrokerClient.subscribedAsyncTopics[topic] {
				subscribedAsyncTopicsByClosedConnection = append(subscribedAsyncTopicsByClosedConnection, topic)
			}
		}
		for topic, _ := range connection.responsibleSyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleSyncTopics, topic)
			if messageBrokerClient.subscribedSyncTopics[topic] {
				subscribedSyncTopicsByClosedConnection = append(subscribedSyncTopicsByClosedConnection, topic)
			}
		}
		delete(messageBrokerClient.brokerConnections, getEndpointString(connection.endpoint))
		messageBrokerClient.mutex.Unlock()

		for _, topic := range subscribedAsyncTopicsByClosedConnection {
			err := messageBrokerClient.startResolutionAttempt(topic, false, stopChannel)

		}
		for _, topic := range subscribedSyncTopicsByClosedConnection {
			err := messageBrokerClient.startResolutionAttempt(topic, true, stopChannel)

		}
	case <-stopChannel:
		messageBrokerClient.mutex.Lock()
		for topic, _ := range connection.responsibleAsyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleAsyncTopics, topic)
		}
		for topic, _ := range connection.responsibleSyncTopics {
			delete(messageBrokerClient.topicResolutions, topic)
			delete(connection.responsibleSyncTopics, topic)
		}
		delete(messageBrokerClient.brokerConnections, getEndpointString(connection.endpoint))
		messageBrokerClient.mutex.Unlock()
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
