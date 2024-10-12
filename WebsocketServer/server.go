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

type websocketServerReceptionManagerCaller struct {
	Client    *WebsocketClient.WebsocketClient
	SessionId string
	Identity  string
}

type WebsocketServer[O any] struct {
	config *Config.WebsocketServer

	name string

	instanceId string
	sessionId  string

	status      int
	statusMutex sync.Mutex
	stopChannel chan struct{}
	waitGroup   sync.WaitGroup

	eventHandler *Event.Handler

	receptionManagerFactory Tools.ReceptionManagerFactory[*websocketServerReceptionManagerCaller]
	acceptionHandler        AcceptionHandler[O]
	//requestResponseManager  *Tools.RequestResponseManager[T]

	websocketListener *WebsocketListener.WebsocketListener

	sessionManager *Tools.SessionManager

	// metrics

	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64

	FailedReceptions     atomic.Uint64
	SuccessfulReceptions atomic.Uint64

	AsyncMessageSent atomic.Uint64
	SyncRequestsSent atomic.Uint64
	SyncResponseSent atomic.Uint64

	AsyncMessageReceived atomic.Uint64
	SyncRequestsReceived atomic.Uint64
	SyncResponseReceived atomic.Uint64

	ClientsAccepted atomic.Uint64
	ClientsRejected atomic.Uint64
}

func New[O any](name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, eventHandleFunc Event.HandleFunc /*  requestResponseManager *Tools.RequestResponseManager[T],  */, acceptionHandler AcceptionHandler[O], receptionManagerFactory Tools.ReceptionManagerFactory[*websocketServerReceptionManagerCaller]) (*WebsocketServer[O], error) {
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

	server := &WebsocketServer[O]{
		config:     config,
		name:       name,
		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		acceptionHandler:        acceptionHandler,
		receptionManagerFactory: receptionManagerFactory,
		//requestResponseManager:  requestResponseManager,
	}
	if server.acceptionHandler == nil {
		server.acceptionHandler = NewDefaultAcceptionHandler[O]()
	}
	if server.receptionManagerFactory == nil {
		server.receptionManagerFactory = NewWebsocketMessageReceptionManagerFactory[O](nil, nil, nil, nil, nil, nil, nil, nil)
	}
	if eventHandleFunc != nil {
		server.eventHandler = Event.NewHandler(eventHandleFunc, server.GetServerContext)
	}
	server.sessionManager = Tools.NewSessionManager(config.SessionManagerConfig, nil, nil)
	websocketListener, err := WebsocketListener.New(server.name+"_websocketListener", server.config.WebsocketListenerConfig, whitelist, blacklist)
	if err != nil {
		return nil, err
	}
	server.websocketListener = websocketListener

	return server, nil
}

func (server *WebsocketServer[O]) GetName() string {
	return server.name
}

func (server *WebsocketServer[O]) GetStatus() int {
	return server.status
}

func (server *WebsocketServer[O]) GetInstanceId() string {
	return server.instanceId
}

func (server *WebsocketServer[O]) GetSessionId() string {
	return server.sessionId
}

func (server *WebsocketServer[O]) GetEventHandler() *Event.Handler {
	return server.eventHandler
}
func (server *WebsocketServer[O]) SetEventHandler(eventHandler Event.HandleFunc) {
	server.eventHandler = Event.NewHandler(eventHandler, server.GetServerContext)
}

func (server *WebsocketServer[O]) SetAcceptionHandler(acceptionHandler AcceptionHandler[O]) {
	server.acceptionHandler = acceptionHandler
}

func (server *WebsocketServer[O]) SetReceptionManagerFactory(receptionManagerFactory Tools.ReceptionManagerFactory[*websocketServerReceptionManagerCaller]) {
	server.receptionManagerFactory = receptionManagerFactory
}

/* func (server *WebsocketServer[T]) GetRequestResponseManager() *Tools.RequestResponseManager[T] {
	return server.requestResponseManager
} */

func (server *WebsocketServer[O]) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:       Event.WebsocketServer,
		Event.ServiceName:       server.name,
		Event.ServiceStatus:     Status.ToString(server.status),
		Event.ServiceInstanceId: server.GetInstanceId(),
		Event.ServiceSessionId:  server.GetSessionId(),
		//Event.Function:          Event.GetCallerFuncName(2),
	}
}
