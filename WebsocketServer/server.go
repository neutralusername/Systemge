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

type AcceptionHandler func(*WebsocketServer, *WebsocketClient.WebsocketClient) (string, error)
type ReceptionHandler func([]byte) error
type ReceptionHandlerFactory func(websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) ReceptionHandler

type ObjectHandler func(object any, websocketServer *WebsocketServer, websocketClient *WebsocketClient.WebsocketClient, identity, sessionId string) error

type WebsocketServer struct {
	config *Config.WebsocketServer

	name string

	instanceId string
	sessionId  string

	status      int
	statusMutex sync.Mutex
	stopChannel chan struct{}
	waitGroup   sync.WaitGroup

	eventHandler *Event.Handler

	getReceptionHandler ReceptionHandlerFactory
	acceptionHandler    AcceptionHandler

	websocketListener *WebsocketListener.WebsocketListener

	sessionManager *Tools.SessionManager

	// metrics

	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64

	MessagesRejected atomic.Uint64
	MessagesAccepted atomic.Uint64

	AsyncMessageSent atomic.Uint64
	SyncRequestsSent atomic.Uint64
	SyncResponseSent atomic.Uint64

	AsyncMessageReceived atomic.Uint64
	SyncRequestsReceived atomic.Uint64
	SyncResponseReceived atomic.Uint64

	ClientsAccepted atomic.Uint64
	ClientsRejected atomic.Uint64
}

func New(name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, acceptionHandler AcceptionHandler, getReceptionHandler ReceptionHandlerFactory, eventHandleFunc Event.HandleFunc) (*WebsocketServer, error) {
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
		config:              config,
		name:                name,
		instanceId:          Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		acceptionHandler:    acceptionHandler,
		getReceptionHandler: getReceptionHandler,
	}
	if server.acceptionHandler == nil {
		server.acceptionHandler = NewDefaultAcceptionHandler()
	}
	if server.getReceptionHandler == nil {
		server.getReceptionHandler = NewDefaultReceptionHandlerFactory()
	}
	if eventHandleFunc != nil {
		server.eventHandler = Event.NewHandler(eventHandleFunc, server.GetServerContext)
	}
	if server.getReceptionHandler == nil {
		server.getReceptionHandler = NewDefaultReceptionHandlerFactory()
	}
	if server.acceptionHandler == nil {
		server.acceptionHandler = NewDefaultAcceptionHandler()
	}
	server.sessionManager = Tools.NewSessionManager(config.SessionManagerConfig, nil, nil)
	websocketListener, err := WebsocketListener.New(server.name+"_websocketListener", server.config.WebsocketListenerConfig, whitelist, blacklist)
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

func (server *WebsocketServer) GetEventHandler() *Event.Handler {
	return server.eventHandler
}
func (server *WebsocketServer) SetEventHandler(eventHandler Event.HandleFunc) {
	server.eventHandler = Event.NewHandler(eventHandler, server.GetServerContext)
}

func (server *WebsocketServer) SetAcceptionHandler(acceptionHandler AcceptionHandler) {
	server.acceptionHandler = acceptionHandler
}

func (server *WebsocketServer) SetGetReceptionHandler(getReceptionHandler ReceptionHandlerFactory) {
	server.getReceptionHandler = getReceptionHandler
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
