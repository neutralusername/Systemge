package Server

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Systemge"
	"github.com/neutralusername/Systemge/Tools"
)

type Server struct {
	instanceId string
	sessionId  string
	name       string

	status      int
	statusMutex sync.RWMutex
	stopChannel chan bool
	waitGroup   sync.WaitGroup

	config *Config.SystemgeServer

	whitelist *Tools.AccessControlList
	blacklist *Tools.AccessControlList

	listener       Systemge.Listener
	sessionManager *SessionManager.Manager

	eventHandler Event.Handler
}

func New(name string, config *Config.SystemgeServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, eventHandler Event.Handler) (*Server, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpSystemgeConnectionConfig == nil {
		return nil, errors.New("config.ConnectionConfig is nil")
	}
	if config.TcpSystemgeListenerConfig == nil {
		return nil, errors.New("config.ListenerConfig is nil")
	}
	if config.TcpSystemgeListenerConfig.TcpServerConfig == nil {
		return nil, errors.New("config.ListenerConfig.ServerConfig is nil")
	}

	server := &Server{
		name:         name,
		config:       config,
		eventHandler: eventHandler,
		instanceId:   Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		whitelist: whitelist,
		blacklist: blacklist,

		sessionManager: SessionManager.New(name+"_sessionManager", config.SessionManagerConfig, eventHandler),
	}
	return server, nil
}

func (server *Server) GetName() string {
	return server.name
}

func (server *Server) GetStatus() int {
	return server.status
}

func (server *Server) GetInstanceId() string {
	return server.instanceId
}

func (server *Server) GetSessionId() string {
	return server.sessionId
}

func (server *Server) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetServerContext())
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *Server) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:       Event.SystemgeServer,
		Event.ServiceName:       server.name,
		Event.ServiceStatus:     Status.ToString(server.status),
		Event.Function:          Event.GetCallerFuncName(2),
		Event.ServiceInstanceId: server.instanceId,
		Event.SessionId:         server.sessionId,
	}
}
