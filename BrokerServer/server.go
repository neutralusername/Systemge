package BrokerServer

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
	name string

	config         *Config.MessageBrokerServer
	systemgeServer *SystemgeServer.SystemgeServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	messageHandlerStopChannel chan<- bool

	messageHandler SystemgeMessageHandler.MessageHandler

	connectionAsyncSubscriptions map[SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true
	connectionsSyncSubscriptions map[SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true

	asyncTopicSubscriptions map[string]map[SystemgeConnection.SystemgeConnection]bool // topic -> connection -> true
	syncTopicSubscriptions  map[string]map[SystemgeConnection.SystemgeConnection]bool // topic -> connection -> true

	mutex sync.Mutex

	// metrics

	asyncMessagesPropagated atomic.Uint64

	syncRequestsPropagated atomic.Uint64
}

func New(name string, config *Config.MessageBrokerServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList) *Server {
	if config == nil {
		panic("config is nil")
	}
	if config.SystemgeServerConfig == nil {
		panic("config.SystemgeServerConfig is nil")
	}
	if config.SystemgeServerConfig.TcpSystemgeListenerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig is nil")
	}
	if config.SystemgeServerConfig.TcpSystemgeListenerConfig.TcpServerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig.ListenerConfig is nil")
	}
	if config.SystemgeServerConfig.TcpSystemgeConnectionConfig == nil {
		panic("config.SystemgeServerConfig.ConnectionConfig is nil")
	}

	server := &Server{
		name:   name,
		config: config,

		asyncTopicSubscriptions: make(map[string]map[SystemgeConnection.SystemgeConnection]bool),
		syncTopicSubscriptions:  make(map[string]map[SystemgeConnection.SystemgeConnection]bool),

		connectionAsyncSubscriptions: make(map[SystemgeConnection.SystemgeConnection]map[string]bool),
		connectionsSyncSubscriptions: make(map[SystemgeConnection.SystemgeConnection]map[string]bool),
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

	server.systemgeServer = SystemgeServer.New(name+"_systemgeServer",
		server.config.SystemgeServerConfig,
		whitelist, blacklist,
		server.onSystemgeConnection, server.onSystemgeDisconnection,
	)

	server.messageHandler = SystemgeMessageHandler.NewSequentialMessageHandler(nil, SystemgeMessageHandler.SyncMessageHandlers{
		Message.TOPIC_SUBSCRIBE_ASYNC:   server.subscribeAsync,
		Message.TOPIC_UNSUBSCRIBE_ASYNC: server.unsubscribeAsync,
		Message.TOPIC_SUBSCRIBE_SYNC:    server.subscribeSync,
		Message.TOPIC_UNSUBSCRIBE_SYNC:  server.unsubscribeSync,
	}, nil, nil, 100000)
	server.AddAsyncTopics(server.config.AsyncTopics)
	server.AddSyncTopics(server.config.SyncTopics)

	return server
}

func (server *Server) Start() error {
	return server.systemgeServer.Start()
}

func (server *Server) Stop() error {
	return server.systemgeServer.Stop()
}

func (server *Server) GetStatus() int {
	return server.systemgeServer.GetStatus()
}

func (server *Server) GetName() string {
	return server.name
}
