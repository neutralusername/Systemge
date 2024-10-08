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

	eventHandler *Event.Handler

	receptionHandler func(*WebsocketClient.WebsocketClient, []byte) error
	acceptionHandler func(*WebsocketClient.WebsocketClient) (string, error)

	websocketListener *WebsocketListener.WebsocketListener

	sessionManager *Tools.SessionManager

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

func New(name string, config *Config.WebsocketServer, acceptionHandler func(*WebsocketClient.WebsocketClient) (string, error), receptionHandler func(*WebsocketClient.WebsocketClient, []byte) error, eventHandler Event.HandleFunc) (*WebsocketServer, error) {
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
		acceptionHandler: acceptionHandler,
		receptionHandler: receptionHandler,
	}
	if eventHandler != nil {
		server.eventHandler = Event.NewHandler(eventHandler, server.GetServerContext)
	}
	if server.receptionHandler == nil {
		server.receptionHandler = GetDefaultReceptionHandler()
	}
	if server.acceptionHandler == nil {
		server.acceptionHandler = GetDefaultAcceptionHandler()
	}
	server.sessionManager = Tools.NewSessionManager(config.SessionManagerConfig, server.onCreateSession, nil)
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

func (server *WebsocketServer) GetEventHandler() *Event.Handler {
	return server.eventHandler
}
func (server *WebsocketServer) SetEventHandler(eventHandler Event.HandleFunc) {
	server.eventHandler = Event.NewHandler(eventHandler, server.GetServerContext)
}

func (server *WebsocketServer) GetAcceptionHandler() func(*WebsocketClient.WebsocketClient) (string, error) {
	return server.acceptionHandler
}
func (server *WebsocketServer) SetAcceptionHandler(acceptionHandler func(*WebsocketClient.WebsocketClient) (string, error)) {
	server.acceptionHandler = acceptionHandler
}

func (server *WebsocketServer) GetReceptionHandler() func(*WebsocketClient.WebsocketClient, []byte) error {
	return server.receptionHandler
}
func (server *WebsocketServer) SetReceptionHandler(receptionHandler func(*WebsocketClient.WebsocketClient, []byte) error) {
	server.receptionHandler = receptionHandler
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
