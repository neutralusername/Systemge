package BrokerServer

import (
	"sync"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
)

type MessageBrokerServer struct {
	config         *Config.MessageBrokerServer
	systemgeServer *SystemgeServer.SystemgeServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	messageHandler  SystemgeConnection.MessageHandler
	dashboardClient *Dashboard.DashboardClient

	asyncConnectionSubscriptions map[*SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true
	syncConnectionSubscriptions  map[*SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true

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

		asyncConnectionSubscriptions: make(map[*SystemgeConnection.SystemgeConnection]map[string]bool),
		syncConnectionSubscriptions:  make(map[*SystemgeConnection.SystemgeConnection]map[string]bool),
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
