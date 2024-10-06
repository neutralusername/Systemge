package WebsocketServer

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketListener"
)

type WebsocketServer struct {
	config *Config.WebsocketServer

	name string

	instanceId string
	sessionId  string

	eventHandler Event.Handler

	websocketListener *WebsocketListener.WebsocketListener

	sessionManager *Tools.SessionManager
	topicManager   *Tools.TopicManager

	status      int
	statusMutex sync.Mutex
	stopChannel chan struct{}
	waitGroup   sync.WaitGroup

	whitelist     *Tools.AccessControlList
	blacklist     *Tools.AccessControlList
	ipRateLimiter *Tools.IpRateLimiter

	// metrics

	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64

	AsyncMessageSent atomic.Uint64
	SyncRequestsSent atomic.Uint64
	SyncResponseSent atomic.Uint64

	AsyncMessageReceived atomic.Uint64
	SyncRequestsReceived atomic.Uint64
	SyncResponseReceived atomic.Uint64
}

func New(name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, eventHandler Event.Handler) (*WebsocketServer, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.SessionManagerConfig == nil {
		return nil, errors.New("config.SessionManagerConfig is nil")
	}
	if config.WebsocketClientConfig == nil {
		return nil, errors.New("config.WebsocketConnectionConfig is nil")
	}
	if config.WebsocketListenerConfig == nil {
		return nil, errors.New("config.WebsocketListenerConfig is nil")
	}

	server := &WebsocketServer{
		name:       name,
		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		whitelist: whitelist,
		blacklist: blacklist,

		config: config,

		eventHandler: eventHandler,
	}
	if config.IpRateLimiterConfig != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(config.IpRateLimiterConfig)
	}
	server.sessionManager = Tools.NewSessionManager(name+"_sessionManager", config.SessionManagerConfig, server.onCreateSession, nil)
	websocketListener, err := WebsocketListener.New(server.name+"_websocketListener", server.config.WebsocketListenerConfig, server.eventHandler)
	if err != nil {
		return nil, err
	}
	server.websocketListener = websocketListener

	return server, nil
}

func (server *WebsocketServer) GetName() string {
	return server.name
}

func (server *WebsocketServer) GetStatus() int {
	return server.status
}

func (server *WebsocketServer) GetInstanceId() string {
	return server.instanceId
}

func (server *WebsocketServer) GetSessionId() string {
	return server.sessionId
}

func (server *WebsocketServer) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetServerContext())
	server.eventHandler(event)
	return event
}
func (server *WebsocketServer) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:   Event.WebsocketServer,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.status),
		//Event.Function:          Event.GetCallerFuncName(2),
		Event.ServiceInstanceId: server.GetInstanceId(),
		Event.ServiceSessionId:  server.GetSessionId(),
	}
}
