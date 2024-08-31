package MessageBroker

import (
	"encoding/json"
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
)

type MessageBrokerServer struct {
	status      int
	statusMutex sync.Mutex

	config         *Config.MessageBrokerServer
	systemgeServer *SystemgeServer.SystemgeServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	messageHandler  SystemgeConnection.MessageHandler
	dashboardClient *Dashboard.DashboardClient

	asyncSubscriptions map[*SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true
	syncSubscriptions  map[*SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true

	asyncTopicSubscriptions map[string]map[*SystemgeConnection.SystemgeConnection]bool // topic -> connection -> true
	syncTopicSubscriptions  map[string]map[*SystemgeConnection.SystemgeConnection]bool // topic -> connection -> true

	mutex sync.Mutex

	// metrics
}

func NewMessageBrokerServer(config *Config.MessageBrokerServer) *MessageBrokerServer {
	if config == nil {
		panic("config is nil")
	}
	if config.SystemgeServerConfig == nil {
		panic("config.SystemgeServerConfig is nil")
	}
	if config.SystemgeServerConfig.ListenerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig is nil")
	}
	if config.SystemgeServerConfig.ListenerConfig.TcpListenerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig.ListenerConfig is nil")
	}
	if config.SystemgeServerConfig.ConnectionConfig == nil {
		panic("config.SystemgeServerConfig.ConnectionConfig is nil")
	}

	server := &MessageBrokerServer{
		config: config,

		asyncTopicSubscriptions: make(map[string]map[*SystemgeConnection.SystemgeConnection]bool),
		syncTopicSubscriptions:  make(map[string]map[*SystemgeConnection.SystemgeConnection]bool),

		asyncSubscriptions: make(map[*SystemgeConnection.SystemgeConnection]map[string]bool),
		syncSubscriptions:  make(map[*SystemgeConnection.SystemgeConnection]map[string]bool),
	}
	if config.InfoLoggerPath != "" {
		server.infoLogger = Tools.NewLogger("[Info: \"MessageBrokerServer\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		server.warningLogger = Tools.NewLogger("[Warning: \"MessageBrokerServer\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		server.errorLogger = Tools.NewLogger("[Error: \"MessageBrokerServer\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		server.mailer = Tools.NewMailer(config.MailerConfig)
	}

	server.systemgeServer = SystemgeServer.New(server.config.SystemgeServerConfig, server.onSystemgeConnection, server.onSystemgeDisconnection)

	if server.config.DashboardClientConfig != nil {
		server.dashboardClient = Dashboard.NewClient(
			server.config.DashboardClientConfig,
			server.systemgeServer.Start, server.systemgeServer.Stop, server.GetMetrics, server.systemgeServer.GetStatus,
			Commands.Handlers{
				Message.TOPIC_SUBSCRIBE_ASYNC: func(args []string) (string, error) {
					server.AddAsyncTopics(args)
					return "success", nil
				},
				Message.TOPIC_SUBSCRIBE_SYNC: func(args []string) (string, error) {
					server.AddSyncTopics(args)
					return "success", nil
				},
				Message.TOPIC_UNSUBSCRIBE_ASYNC: func(args []string) (string, error) {
					server.RemoveAsyncTopics(args)
					return "success", nil
				},
				Message.TOPIC_UNSUBSCRIBE_SYNC: func(args []string) (string, error) {
					server.RemoveSyncTopics(args)
					return "success", nil
				},
			},
		)

		if err := server.StartDashboardClient(); err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Error.New("failed to start dashboard client", err).Error())
			}
			if server.mailer != nil {
				if err := server.mailer.Send(Tools.NewMail(nil, "error", Error.New("failed to start dashboard client", err).Error())); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("failed to send mail", err).Error())
					}
				}
			}
		}
	}
	server.messageHandler = SystemgeConnection.NewSequentialMessageHandler(nil, SystemgeConnection.SyncMessageHandlers{
		Message.TOPIC_SUBSCRIBE_ASYNC:   server.subscribeAsync,
		Message.TOPIC_UNSUBSCRIBE_ASYNC: server.unsubscribeAsync,
		Message.TOPIC_SUBSCRIBE_SYNC:    server.subscribeSync,
		Message.TOPIC_UNSUBSCRIBE_SYNC:  server.unsubscribeSync,
	}, nil, nil, 100000)
	server.AddAsyncTopics(server.config.AsyncTopics)
	server.AddSyncTopics(server.config.SyncTopics)
	return server
}

func (server *MessageBrokerServer) StartDashboardClient() error {
	if server.dashboardClient == nil {
		return Error.New("dashboard client is not enabled", nil)
	}
	return server.dashboardClient.Start()
}

func (server *MessageBrokerServer) StopDashboardClient() error {
	if server.dashboardClient == nil {
		return Error.New("dashboard client is not enabled", nil)
	}
	return server.dashboardClient.Stop()
}

func (server *MessageBrokerServer) Start() error {
	return server.systemgeServer.Start()
}

func (server *MessageBrokerServer) Stop() error {
	return server.systemgeServer.Stop()
}

func (server *MessageBrokerServer) GetStatus() int {
	return server.systemgeServer.GetStatus()
}

func (server *MessageBrokerServer) GetMetrics() map[string]uint64 {
	// TODO: gather metrics
	return nil
}

func (server *MessageBrokerServer) AddAsyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.messageHandler.AddAsyncMessageHandler(topic, server.handleAsyncPropagate)
		if server.asyncTopicSubscriptions[topic] == nil {
			server.asyncTopicSubscriptions[topic] = make(map[*SystemgeConnection.SystemgeConnection]bool)
		}
	}
}

func (server *MessageBrokerServer) AddSyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.messageHandler.AddSyncMessageHandler(topic, server.handleSyncPropagate)
		if server.syncTopicSubscriptions[topic] == nil {
			server.syncTopicSubscriptions[topic] = make(map[*SystemgeConnection.SystemgeConnection]bool)
		}
	}
}

func (server *MessageBrokerServer) RemoveAsyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicSubscriptions[topic] != nil {
			server.messageHandler.RemoveAsyncMessageHandler(topic)
			delete(server.asyncTopicSubscriptions, topic)
		}
	}
}

func (server *MessageBrokerServer) RemoveSyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicSubscriptions[topic] != nil {
			server.messageHandler.RemoveSyncMessageHandler(topic)
			delete(server.syncTopicSubscriptions, topic)
		}
	}
}

func (server *MessageBrokerServer) subscribeAsync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicSubscriptions[topic] == nil {
			return "", Error.New("unknown async topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		server.asyncTopicSubscriptions[topic][connection] = true
		server.asyncSubscriptions[connection][topic] = true
	}
	return "", nil
}

func (server *MessageBrokerServer) subscribeSync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicSubscriptions[topic] == nil {
			return "", Error.New("unknown sync topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		server.syncTopicSubscriptions[topic][connection] = true
		server.syncSubscriptions[connection][topic] = true
	}
	return "", nil
}

func (server *MessageBrokerServer) unsubscribeAsync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicSubscriptions[topic] == nil {
			return "", Error.New("unknown async topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		delete(server.asyncTopicSubscriptions[topic], connection)
		delete(server.asyncSubscriptions[connection], topic)
	}
	return "", nil
}

func (server *MessageBrokerServer) unsubscribeSync(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicSubscriptions[topic] == nil {
			return "", Error.New("unknown sync topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		delete(server.syncTopicSubscriptions[topic], connection)
		delete(server.syncSubscriptions[connection], topic)
	}
	return "", nil
}

func (server *MessageBrokerServer) handleAsyncPropagate(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for client := range server.asyncTopicSubscriptions[message.GetTopic()] {
		if client != connection {
			client.AsyncMessage(message.GetTopic(), message.GetPayload())
		}
	}
}

func (server *MessageBrokerServer) handleSyncPropagate(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	responseChannels := []<-chan *Message.Message{}
	for client := range server.syncTopicSubscriptions[message.GetTopic()] {
		if client != connection {
			responseChannel, err := client.SyncRequest(message.GetTopic(), message.GetPayload())
			if err != nil {
				if server.warningLogger != nil {
					server.warningLogger.Log(Error.New("failed to send sync request to client \""+client.GetName(), nil).Error())
				}
				continue
			}
			responseChannels = append(responseChannels, responseChannel)
		}
	}
	responses := []string{}
	for _, responseChannel := range responseChannels {
		response := <-responseChannel
		if response != nil {
			responses = append(responses, string(response.Serialize()))
		}
	}
	if len(responses) == 0 {
		return "", Error.New("no responses", nil)
	}
	return Helpers.StringsToJsonObjectArray(responses), nil
}

func (server *MessageBrokerServer) onSystemgeConnection(connection *SystemgeConnection.SystemgeConnection) error {
	server.mutex.Lock()
	server.asyncSubscriptions[connection] = make(map[string]bool)
	server.syncSubscriptions[connection] = make(map[string]bool)
	server.mutex.Unlock()
	connection.StartProcessingLoopSequentially(server.messageHandler)
	return nil
}

func (server *MessageBrokerServer) onSystemgeDisconnection(connection *SystemgeConnection.SystemgeConnection) {
	connection.StopProcessingLoop()
	server.mutex.Lock()
	for topic := range server.asyncSubscriptions[connection] {
		delete(server.asyncTopicSubscriptions[topic], connection)
	}
	delete(server.asyncSubscriptions, connection)
	for topic := range server.syncSubscriptions[connection] {
		delete(server.syncTopicSubscriptions[topic], connection)
	}
	delete(server.syncSubscriptions, connection)
	server.mutex.Unlock()
}
