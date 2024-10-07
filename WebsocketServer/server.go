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
	"github.com/neutralusername/Systemge/WebsocketClient"
	"github.com/neutralusername/Systemge/WebsocketListener"
)

type WebsocketServer struct {
	config *Config.WebsocketServer

	name string

	instanceId string
	sessionId  string

	status      int
	statusMutex sync.Mutex
	stopChannel chan struct{}
	waitGroup   sync.WaitGroup

	eventHandler Event.Handler

	messageHandler   func(*WebsocketClient.WebsocketClient, []byte) error
	handshakeHandler func(*WebsocketClient.WebsocketClient) (string, error)

	whitelist     *Tools.AccessControlList
	blacklist     *Tools.AccessControlList
	ipRateLimiter *Tools.IpRateLimiter

	websocketListener *WebsocketListener.WebsocketListener

	sessionManager *Tools.SessionManager
	topicManager   *Tools.TopicManager

	// metrics

	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64

	RejectedMessages atomic.Uint64
	AcceptedMessages atomic.Uint64

	AsyncMessageSent atomic.Uint64
	SyncRequestsSent atomic.Uint64
	SyncResponseSent atomic.Uint64

	AsyncMessageReceived atomic.Uint64
	SyncRequestsReceived atomic.Uint64
	SyncResponseReceived atomic.Uint64

	ClientsAccepted atomic.Uint64
	ClientsRejected atomic.Uint64
}

func New(name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, handshakeHandler func(*WebsocketClient.WebsocketClient) (string, error), messageHandler func(*WebsocketClient.WebsocketClient, []byte) error, eventHandler Event.Handler) (*WebsocketServer, error) {
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
		config:           config,
		name:             name,
		instanceId:       Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		eventHandler:     eventHandler,
		handshakeHandler: handshakeHandler,
		messageHandler:   messageHandler,
		whitelist:        whitelist,
		blacklist:        blacklist,
	}
	if server.messageHandler == nil {
		server.messageHandler = server.GetDefaultMessageHandler()
	}
	server.sessionManager = Tools.NewSessionManager(config.SessionManagerConfig, server.onCreateSession, nil)
	if config.IpRateLimiterConfig != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(config.IpRateLimiterConfig)
	}
	websocketListener, err := WebsocketListener.New(server.name+"_websocketListener", server.config.WebsocketListenerConfig)
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
		Event.ServiceType:       Event.WebsocketServer,
		Event.ServiceName:       server.name,
		Event.ServiceStatus:     Status.ToString(server.status),
		Event.ServiceInstanceId: server.GetInstanceId(),
		Event.ServiceSessionId:  server.GetSessionId(),
		//Event.Function:          Event.GetCallerFuncName(2),
	}
}
