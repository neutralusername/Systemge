package BrokerClient

import (
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeClient"
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

	waitGroup   sync.WaitGroup
	stopChannel chan bool

	messageHandler SystemgeConnection.MessageHandler

	subscribedTopicSystemgeClient *SystemgeClient.SystemgeClient
	topicResolutionSystemgeClient *SystemgeClient.SystemgeClient

	ongoingTopicResolutions     map[string]*resolutionAttempt
	ongoingGetBrokerConnections map[string]*getBrokerConnectionAttempt

	topicResolutions       map[string]map[string]SystemgeConnection.SystemgeConnection // topic -> [endpointString -> systemgeConnection]
	responsibleAsyncTopics map[SystemgeConnection.MessageHandler]bool
	responsibleSyncTopics  map[SystemgeConnection.MessageHandler]bool

	mutex sync.Mutex

	subscribedAsyncTopics map[string]bool
	subscribedSyncTopics  map[string]bool

	// metrics

	asyncMessagesSent atomic.Uint64

	syncRequestsSent      atomic.Uint64
	syncResponsesReceived atomic.Uint64

	resolutionAttempts atomic.Uint64
}

func New(name string, config *Config.MessageBrokerClient, systemgeMessageHandler SystemgeConnection.MessageHandler, dashboardCommands Commands.Handlers) *Client {
	if config == nil {
		panic(Error.New("Config is required", nil))
	}
	if config.ResolverTcpSystemgeConnectionConfig == nil {
		panic(Error.New("ResolverConnectionConfig is required", nil))
	}
	if config.BrokerSystemgeClientConfig == nil {
		panic(Error.New("ConnectionConfig is required", nil))
	}
	if len(config.ResolverTcpClientConfigs) == 0 {
		panic(Error.New("At least one ResolverEndpoint is required", nil))
	}

	messageBrokerClient := &Client{
		name:                    name,
		config:                  config,
		messageHandler:          systemgeMessageHandler,
		ongoingTopicResolutions: make(map[string]*resolutionAttempt),

		topicResolutions:       make(map[string]map[string]SystemgeConnection.SystemgeConnection),
		responsibleAsyncTopics: make(map[SystemgeConnection.MessageHandler]bool),
		responsibleSyncTopics:  make(map[SystemgeConnection.MessageHandler]bool),

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

	messageBrokerClient.subscribedTopicSystemgeClient = SystemgeClient.New(name,
		config.BrokerSystemgeClientConfig,
		func(systemgeConnection SystemgeConnection.SystemgeConnection) error {
			connections := map[string]*connection{}
			for _, endpoint := range endpoints {
				conn, err := messageBrokerClient.getBrokerConnection(endpoint, stopChannel)
				if err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Error.New("Failed to get connection to resolved endpoint \""+endpoint.Address+"\" for topic \""+resolutionAttempt.topic+"\"", err).Error())
					}
					if messageBrokerClient.mailer != nil {
						if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to get connection to resolved endpoint \""+endpoint.Address+"\" for topic \""+resolutionAttempt.topic+"\"", err).Error())); err != nil {
							if messageBrokerClient.errorLogger != nil {
								messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
							}
						}
					}
					continue
				}
				messageBrokerClient.mutex.Lock()
				if resolutionAttempt.isSyncTopic {
					conn.responsibleSyncTopics[resolutionAttempt.topic] = true
				} else {
					conn.responsibleAsyncTopics[resolutionAttempt.topic] = true
				}
				connections[getEndpointString(endpoint)] = conn
				messageBrokerClient.mutex.Unlock()

				if subscribe {
					messageBrokerClient.subscribeToTopic(conn, resolutionAttempt.topic, resolutionAttempt.isSyncTopic)
				}
			}

			messageBrokerClient.mutex.Lock()
			for endpointString, existingConnection := range messageBrokerClient.topicResolutions[resolutionAttempt.topic] {
				if _, ok := connections[endpointString]; !ok {
					if resolutionAttempt.isSyncTopic {
						delete(existingConnection.responsibleSyncTopics, resolutionAttempt.topic)
					} else {
						delete(existingConnection.responsibleAsyncTopics, resolutionAttempt.topic)
					}
					if len(existingConnection.responsibleAsyncTopics) == 0 && len(existingConnection.responsibleSyncTopics) == 0 {
						delete(messageBrokerClient.brokerConnections, endpointString)
						existingConnection.connection.Close()
					}
				}
			}
			messageBrokerClient.topicResolutions[resolutionAttempt.topic] = connections
			resolutionAttempt.connections = connections

			delete(messageBrokerClient.ongoingTopicResolutions, resolutionAttempt.topic)
			close(resolutionAttempt.ongoing)
			messageBrokerClient.mutex.Unlock()

			go messageBrokerClient.handleTopicResolutionLifetime(resolutionAttempt.topic, resolutionAttempt.isSyncTopic, stopChannel)
			return nil
		},
		func(systemgeConnection SystemgeConnection.SystemgeConnection) {

		},
	)

	for _, asyncTopic := range config.AsyncTopics {
		messageBrokerClient.subscribedAsyncTopics[asyncTopic] = true
		messageBrokerClient.topicResolutions[asyncTopic] = make(map[string]SystemgeConnection.SystemgeConnection)
	}
	for _, syncTopic := range config.SyncTopics {
		messageBrokerClient.subscribedSyncTopics[syncTopic] = true
		messageBrokerClient.topicResolutions[syncTopic] = make(map[string]SystemgeConnection.SystemgeConnection)
	}
	return messageBrokerClient
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
		resolutionAttempt, _ := messageBrokerClient.startResolutionAttempt(topic, false)
		<-resolutionAttempt.ongoing
		messageBrokerClient.addSubscribedTopicConnectionAttempts(resolutionAttempt.endpoints, topic)
	}
	for topic := range messageBrokerClient.subscribedSyncTopics {
		resolutionAttempt, _ := messageBrokerClient.startResolutionAttempt(topic, true)
		<-resolutionAttempt.ongoing
		messageBrokerClient.addSubscribedTopicConnectionAttempts(resolutionAttempt.endpoints, topic)
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

func (messageBrokerClient *Client) GetName() string {
	return messageBrokerClient.name
}

func getEndpointString(endpoint *Config.TcpClient) string {
	return endpoint.Address + endpoint.TlsCert
}

func (messageBrokerClient *Client) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{
		"start": func(args []string) (string, error) {
			err := messageBrokerClient.Start()
			if err != nil {
				return "", Error.New("Failed to start message broker client", err)
			}
			return "success", nil
		},
		"stop": func(args []string) (string, error) {
			err := messageBrokerClient.Stop()
			if err != nil {
				return "", Error.New("Failed to stop message broker client", err)
			}
			return "success", nil
		},
		"getStatus": func(args []string) (string, error) {
			return Status.ToString(messageBrokerClient.GetStatus()), nil
		},
		"checkMetrics": func(args []string) (string, error) {
			metrics := messageBrokerClient.CheckMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Error.New("Failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"getMetrics": func(args []string) (string, error) {
			metrics := messageBrokerClient.GetMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Error.New("Failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"resolveTopic": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.ResolveTopic(args[0])
		},
		"resolveSubscribeTopics": func(args []string) (string, error) {
			return "success", messageBrokerClient.ResolveSubscribeTopics()
		},
		"getAsyncSubscribeTopics": func(args []string) (string, error) {
			topics := messageBrokerClient.GetAsyncSubscribeTopics()
			return Helpers.JsonMarshal(topics), nil
		},
		"getSyncSubscribeTopics": func(args []string) (string, error) {
			topics := messageBrokerClient.GetSyncSubscribeTopics()
			return Helpers.JsonMarshal(topics), nil
		},
		"addAsyncSubscribeTopic": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.AddAsyncSubscribeTopic(args[0])
		},
		"addSyncSubscribeTopic": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.AddSyncSubscribeTopic(args[0])
		},
		"removeAsyncSubscribeTopic": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.RemoveAsyncSubscribeTopic(args[0])
		},
		"removeSyncSubscribeTopic": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.RemoveSyncSubscribeTopic(args[0])
		},
		"asyncMessage": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			topic := args[0]
			payload := args[1]
			messageBrokerClient.AsyncMessage(topic, payload)
			return "success", nil
		},
		"syncRequest": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Error.New("Invalid number of arguments", nil)
			}
			topic := args[0]
			payload := args[1]
			responseMessages := messageBrokerClient.SyncRequest(topic, payload)
			json, err := json.Marshal(responseMessages)
			if err != nil {
				return "", Error.New("Failed to marshal messages to json", err)
			}
			return string(json), nil
		},
	}

	return commands
}
