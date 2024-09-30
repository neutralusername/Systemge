package BrokerServer

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
	name string

	config         *Config.MessageBrokerServer
	systemgeServer *SystemgeServer.SystemgeServer

	eventHandler Event.Handler

	messageHandler SystemgeConnection.MessageHandler

	connectionAsyncSubscriptions map[SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true
	connectionsSyncSubscriptions map[SystemgeConnection.SystemgeConnection]map[string]bool // connection -> topic -> true

	asyncTopicSubscriptions map[string]map[SystemgeConnection.SystemgeConnection]bool // topic -> connection -> true
	syncTopicSubscriptions  map[string]map[SystemgeConnection.SystemgeConnection]bool // topic -> connection -> true

	mutex sync.Mutex

	// metrics

	asyncMessagesPropagated atomic.Uint64

	syncRequestsPropagated atomic.Uint64
}

func New(name string, config *Config.MessageBrokerServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, eventHandler Event.Handler) (*Server, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.SystemgeServerConfig == nil {
		return nil, errors.New("config.SystemgeServerConfig is nil")
	}
	if config.SystemgeServerConfig.TcpSystemgeListenerConfig == nil {
		return nil, errors.New("config.SystemgeServerConfig.ListenerConfig is nil")
	}
	if config.SystemgeServerConfig.TcpSystemgeListenerConfig.TcpServerConfig == nil {
		return nil, errors.New("config.SystemgeServerConfig.ServerConfig.ListenerConfig is nil")
	}
	if config.SystemgeServerConfig.TcpSystemgeConnectionConfig == nil {
		return nil, errors.New("config.SystemgeServerConfig.ConnectionConfig is nil")
	}

	server := &Server{
		name:   name,
		config: config,

		asyncTopicSubscriptions: make(map[string]map[SystemgeConnection.SystemgeConnection]bool),
		syncTopicSubscriptions:  make(map[string]map[SystemgeConnection.SystemgeConnection]bool),

		connectionAsyncSubscriptions: make(map[SystemgeConnection.SystemgeConnection]map[string]bool),
		connectionsSyncSubscriptions: make(map[SystemgeConnection.SystemgeConnection]map[string]bool),
	}

	systemgeServer, err := SystemgeServer.New(name+"_systemgeServer",
		server.config.SystemgeServerConfig,
		whitelist, blacklist,
		eventHandler, // wrap the onConnect/onDisconnect with the eventHandler
	)
	if err != nil {
		return nil, err
	}
	server.systemgeServer = systemgeServer

	server.messageHandler = SystemgeConnection.NewTopicExclusiveMessageHandler(
		nil,
		SystemgeConnection.SyncMessageHandlers{
			Message.TOPIC_SUBSCRIBE_ASYNC:   server.subscribeAsync,
			Message.TOPIC_UNSUBSCRIBE_ASYNC: server.unsubscribeAsync,
			Message.TOPIC_SUBSCRIBE_SYNC:    server.subscribeSync,
			Message.TOPIC_UNSUBSCRIBE_SYNC:  server.unsubscribeSync,
		},
		nil, nil,
		server.config.MessageHandlerQueueSize,
	)
	server.AddAsyncTopics(server.config.AsyncTopics)
	server.AddSyncTopics(server.config.SyncTopics)

	return server, nil
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
