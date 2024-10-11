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

type structName123 struct {
}

type WebsocketServer[T any] struct {
	config *Config.WebsocketServer

	name string

	instanceId string
	sessionId  string

	status      int
	statusMutex sync.Mutex
	stopChannel chan struct{}
	waitGroup   sync.WaitGroup

	eventHandler *Event.Handler

	receptionHandlerFactory Tools.ReceptionHandlerFactory[*structName123]
	acceptionHandler        AcceptionHandler[T]
	//requestResponseManager  *Tools.RequestResponseManager[T]

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

func New[T any](name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, eventHandleFunc Event.HandleFunc /*  requestResponseManager *Tools.RequestResponseManager[T],  */, acceptionHandler AcceptionHandler[T], receptionHandlerFactory Tools.ReceptionHandlerFactory[*structName123]) (*WebsocketServer[T], error) {
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
	/* if requestResponseManager == nil {
		return nil, errors.New("requestResponseManager is nil")
	} */

	server := &WebsocketServer[T]{
		config:     config,
		name:       name,
		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		acceptionHandler:        acceptionHandler,
		receptionHandlerFactory: receptionHandlerFactory,
		//requestResponseManager:  requestResponseManager,
	}
	if server.acceptionHandler == nil {
		server.acceptionHandler = NewDefaultAcceptionHandler[T]()
	}
	if server.receptionHandlerFactory == nil {
		server.receptionHandlerFactory = NewDefaultReceptionHandlerFactory[T]()
	}
	if eventHandleFunc != nil {
		server.eventHandler = Event.NewHandler(eventHandleFunc, server.GetServerContext)
	}
	if server.receptionHandlerFactory == nil {
		server.receptionHandlerFactory = NewDefaultReceptionHandlerFactory[T]()
	}
	if server.acceptionHandler == nil {
		server.acceptionHandler = NewDefaultAcceptionHandler[T]()
	}
	server.sessionManager = Tools.NewSessionManager(config.SessionManagerConfig, nil, nil)
	websocketListener, err := WebsocketListener.New(server.name+"_websocketListener", server.config.WebsocketListenerConfig, whitelist, blacklist)
	if err != nil {
		return nil, err
	}
	server.websocketListener = websocketListener

	return server, nil
}

func (server *WebsocketServer[T]) GetName() string {
	return server.name
}

func (server *WebsocketServer[T]) GetStatus() int {
	return server.status
}

func (server *WebsocketServer[T]) GetInstanceId() string {
	return server.instanceId
}

func (server *WebsocketServer[T]) GetSessionId() string {
	return server.sessionId
}

func (server *WebsocketServer[T]) GetEventHandler() *Event.Handler {
	return server.eventHandler
}
func (server *WebsocketServer[T]) SetEventHandler(eventHandler Event.HandleFunc) {
	server.eventHandler = Event.NewHandler(eventHandler, server.GetServerContext)
}

func (server *WebsocketServer[T]) SetAcceptionHandler(acceptionHandler AcceptionHandler[T]) {
	server.acceptionHandler = acceptionHandler
}

func (server *WebsocketServer[T]) SetGetReceptionHandler(getReceptionHandler WebsocketServerReceptionHandlerFactory[T]) {
	server.receptionHandlerFactory = getReceptionHandler
}

/* func (server *WebsocketServer[T]) GetRequestResponseManager() *Tools.RequestResponseManager[T] {
	return server.requestResponseManager
} */

func (server *WebsocketServer[T]) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:       Event.WebsocketServer,
		Event.ServiceName:       server.name,
		Event.ServiceStatus:     Status.ToString(server.status),
		Event.ServiceInstanceId: server.GetInstanceId(),
		Event.ServiceSessionId:  server.GetSessionId(),
		//Event.Function:          Event.GetCallerFuncName(2),
	}
}
