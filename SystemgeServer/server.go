package SystemgeServer

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeListener"
	"github.com/neutralusername/Systemge/Tools"
)

type OnConnectHandler func(SystemgeConnection.SystemgeConnection) error
type OnDisconnectHandler func(string, string)

type SystemgeServer struct {
	instanceId string
	sessionId  string
	name       string

	status      int
	statusMutex sync.RWMutex

	whitelist *Tools.AccessControlList
	blacklist *Tools.AccessControlList

	config   *Config.SystemgeServer
	listener SystemgeListener.SystemgeListener

	onConnectHandler    func(SystemgeConnection.SystemgeConnection) error
	onDisconnectHandler func(SystemgeConnection.SystemgeConnection)

	clients     map[string]SystemgeConnection.SystemgeConnection // name -> connection
	mutex       sync.Mutex
	stopChannel chan bool

	waitGroup sync.WaitGroup

	eventHandler Event.Handler
}

func New(name string, config *Config.SystemgeServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, onConnectHandler func(SystemgeConnection.SystemgeConnection) error, onDisconnectHandler func(SystemgeConnection.SystemgeConnection)) *SystemgeServer {
	if config == nil {
		panic("config is nil")
	}
	if config.TcpSystemgeConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.TcpSystemgeListenerConfig == nil {
		panic("listener is nil")
	}
	if config.TcpSystemgeListenerConfig.TcpServerConfig == nil {
		panic("listener.ListenerConfig is nil")
	}

	server := &SystemgeServer{
		name:                name,
		config:              config,
		onConnectHandler:    onConnectHandler,
		onDisconnectHandler: onDisconnectHandler,
		instanceId:          Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		whitelist: whitelist,
		blacklist: blacklist,

		clients: make(map[string]SystemgeConnection.SystemgeConnection),
	}
	return server
}

func (server *SystemgeServer) GetName() string {
	return server.name
}

func (server *SystemgeServer) GetStatus() int {
	return server.status
}

func (server *SystemgeServer) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetServerContext())
	if server.eventHandler == nil {
		return event
	}
	return server.eventHandler(event)
}
func (server *SystemgeServer) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.WebsocketServer,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.status),
		Event.Function:      Event.GetCallerFuncName(2),
		Event.InstanceId:    server.instanceId,
		Event.SessionId:     server.sessionId,
	}
}
