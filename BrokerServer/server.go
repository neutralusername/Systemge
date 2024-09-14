package BrokerServer

import (
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
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

	messageHandler SystemgeConnection.MessageHandler

	connectionAsyncSubscriptions map[SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true
	connectionsSyncSubscriptions map[SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true

	asyncTopicSubscriptions map[string]map[SystemgeConnection.SystemgeConnection]bool // topic -> connection -> true
	syncTopicSubscriptions  map[string]map[SystemgeConnection.SystemgeConnection]bool // topic -> connection -> true

	mutex sync.Mutex

	// metrics

	asyncMessagesReceived   atomic.Uint64
	asyncMessagesPropagated atomic.Uint64

	syncRequestsReceived   atomic.Uint64
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

func (server *Server) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{
		"start": func(args []string) (string, error) {
			err := server.Start()
			if err != nil {
				return "", Error.New("failed to start message broker server", err)
			}
			return "success", nil
		},
		"stop": func(args []string) (string, error) {
			err := server.Stop()
			if err != nil {
				return "", Error.New("failed to stop message broker server", err)
			}
			return "success", nil
		},
		"getStatus": func(args []string) (string, error) {
			return Status.ToString(server.GetStatus()), nil
		},
		"getMetrics": func(args []string) (string, error) {
			metrics := server.GetMetrics_()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Error.New("failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"retrieveMetrics": func(args []string) (string, error) {
			metrics := server.GetMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Error.New("failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"addAsyncTopics": func(args []string) (string, error) {
			server.AddAsyncTopics(args)
			return "success", nil
		},
		"removeAsyncTopics": func(args []string) (string, error) {
			server.RemoveAsyncTopics(args)
			return "success", nil
		},
		"addSyncTopics": func(args []string) (string, error) {
			server.AddSyncTopics(args)
			return "success", nil
		},
		"removeSyncTopics": func(args []string) (string, error) {
			server.RemoveSyncTopics(args)
			return "success", nil
		},
	}
	systemgeServerCommands := server.systemgeServer.GetDefaultCommands()
	for key, value := range systemgeServerCommands {
		commands["systemgeServer_"+key] = value
	}
	return commands
}
