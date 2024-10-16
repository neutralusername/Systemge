package BrokerServer

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	Server1 "github.com/neutralusername/Systemge/Server"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
	name string

	config         *Config.MessageBrokerServer
	systemgeServer *Server1.Server

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

	systemgeServer, err := Server1.New(name+"_systemgeServer",
		server.config.SystemgeServerConfig,
		whitelist, blacklist,
		func(event *Event.Event) {
			if eventHandler != nil {
				eventHandler(event)
			}

			switch event.GetEvent() {
			case Event.HandledAcception:
				server.mutex.Lock()
				err := connection.StartMessageHandlingLoop(server.messageHandler, true)
				if err != nil {
					server.mutex.Unlock()
					return err
				}
				server.connectionAsyncSubscriptions[connection] = make(map[string]bool)
				server.connectionsSyncSubscriptions[connection] = make(map[string]bool)
				server.mutex.Unlock()

			case Event.HandledDisconnection:
				server.mutex.Lock()
				for topic := range server.connectionAsyncSubscriptions[connection] {
					delete(server.asyncTopicSubscriptions[topic], connection)
				}
				delete(server.connectionAsyncSubscriptions, connection)
				for topic := range server.connectionsSyncSubscriptions[connection] {
					delete(server.syncTopicSubscriptions[topic], connection)
				}
				delete(server.connectionsSyncSubscriptions, connection)
				server.mutex.Unlock()
			}
		},
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
