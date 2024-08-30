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

	clientsAsyncTopics map[*SystemgeConnection.SystemgeConnection]map[string]bool
	clientsSyncTopics  map[*SystemgeConnection.SystemgeConnection]map[string]bool

	asyncTopicResolutions map[string]map[*SystemgeConnection.SystemgeConnection]bool
	syncTopicResolutions  map[string]map[*SystemgeConnection.SystemgeConnection]bool

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

		asyncTopicResolutions: make(map[string]map[*SystemgeConnection.SystemgeConnection]bool),
		syncTopicResolutions:  make(map[string]map[*SystemgeConnection.SystemgeConnection]bool),

		clientsAsyncTopics: make(map[*SystemgeConnection.SystemgeConnection]map[string]bool),
		clientsSyncTopics:  make(map[*SystemgeConnection.SystemgeConnection]map[string]bool),
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
				Message.TOPIC_ADD_ASYNC_TOPICS: func(args []string) (string, error) {
					server.AddAsyncTopics(args)
					return "success", nil
				},
				Message.TOPIC_ADD_SYNC_TOPICS: func(args []string) (string, error) {
					server.AddSyncTopics(args)
					return "success", nil
				},
				Message.TOPIC_REMOVE_ASYNC_TOPICS: func(args []string) (string, error) {
					server.RemoveAsyncTopics(args)
					return "success", nil
				},
				Message.TOPIC_REMOVE_SYNC_TOPICS: func(args []string) (string, error) {
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
		Message.TOPIC_ADD_ASYNC_TOPICS:    server.addAsyncTopics,
		Message.TOPIC_REMOVE_ASYNC_TOPICS: server.removeAsyncTopics,
		Message.TOPIC_ADD_SYNC_TOPICS:     server.addSyncTopics,
		Message.TOPIC_REMOVE_SYNC_TOPICS:  server.removeSyncTopics,
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
		if server.asyncTopicResolutions[topic] == nil {
			server.asyncTopicResolutions[topic] = make(map[*SystemgeConnection.SystemgeConnection]bool)
		}
	}
}

func (server *MessageBrokerServer) AddSyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		server.messageHandler.AddSyncMessageHandler(topic, server.handleSyncPropagate)
		if server.syncTopicResolutions[topic] == nil {
			server.syncTopicResolutions[topic] = make(map[*SystemgeConnection.SystemgeConnection]bool)
		}
	}
}

func (server *MessageBrokerServer) RemoveAsyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicResolutions[topic] != nil {
			server.messageHandler.RemoveAsyncMessageHandler(topic)
			delete(server.asyncTopicResolutions, topic)
		}
	}
}

func (server *MessageBrokerServer) RemoveSyncTopics(topics []string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicResolutions[topic] != nil {
			server.messageHandler.RemoveSyncMessageHandler(topic)
			delete(server.syncTopicResolutions, topic)
		}
	}
}

func (server *MessageBrokerServer) addAsyncTopics(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicResolutions[topic] == nil {
			return "", Error.New("unknown async topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		server.asyncTopicResolutions[topic][connection] = true
		server.clientsAsyncTopics[connection][topic] = true
	}
	return "", nil
}

func (server *MessageBrokerServer) addSyncTopics(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicResolutions[topic] == nil {
			return "", Error.New("unknown sync topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		server.syncTopicResolutions[topic][connection] = true
		server.clientsSyncTopics[connection][topic] = true
	}
	return "", nil
}

func (server *MessageBrokerServer) removeAsyncTopics(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.asyncTopicResolutions[topic] == nil {
			return "", Error.New("unknown async topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		delete(server.asyncTopicResolutions[topic], connection)
		delete(server.clientsAsyncTopics[connection], topic)
	}
	return "", nil
}

func (server *MessageBrokerServer) removeSyncTopics(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	topics := []string{}
	err := json.Unmarshal([]byte(message.GetPayload()), &topics)
	if err != nil {
		return "", err
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		if server.syncTopicResolutions[topic] == nil {
			return "", Error.New("unknown sync topic \""+topic+"\"", nil)
		}
	}
	for _, topic := range topics {
		delete(server.syncTopicResolutions[topic], connection)
		delete(server.clientsSyncTopics[connection], topic)
	}
	return "", nil
}

func (server *MessageBrokerServer) handleAsyncPropagate(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for client := range server.asyncTopicResolutions[message.GetTopic()] {
		if client != connection {
			client.AsyncMessage(message.GetTopic(), message.GetPayload())
		}
	}
}

func (server *MessageBrokerServer) handleSyncPropagate(connection *SystemgeConnection.SystemgeConnection, message *Message.Message) (string, error) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	responseChannels := []<-chan *Message.Message{}
	for client := range server.syncTopicResolutions[message.GetTopic()] {
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
	server.clientsAsyncTopics[connection] = make(map[string]bool)
	server.clientsSyncTopics[connection] = make(map[string]bool)
	server.mutex.Unlock()
	connection.StartProcessingLoopSequentially(server.messageHandler)
	return nil
}

func (server *MessageBrokerServer) onSystemgeDisconnection(connection *SystemgeConnection.SystemgeConnection) {
	connection.StopProcessingLoop()
	server.mutex.Lock()
	for topic := range server.clientsAsyncTopics[connection] {
		delete(server.asyncTopicResolutions[topic], connection)
	}
	delete(server.clientsAsyncTopics, connection)
	for topic := range server.clientsSyncTopics[connection] {
		delete(server.syncTopicResolutions[topic], connection)
	}
	delete(server.clientsSyncTopics, connection)
	server.mutex.Unlock()
}
