package WebsocketServer

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketServer struct {
	name string

	status      int
	statusMutex sync.Mutex

	stopChannel chan bool
	waitGroup   sync.WaitGroup

	instanceId string
	sessionId  string

	config *Config.WebsocketServer

	httpServer        *HTTPServer.HTTPServer
	connectionChannel chan *websocket.Conn
	ipRateLimiter     *Tools.IpRateLimiter

	sessionManager *Tools.SessionManager

	topicManager *Tools.TopicManager

	eventHandler Event.Handler

	// metrics

	websocketConnectionsAccepted atomic.Uint32
	websocketConnectionsFailed   atomic.Uint32
	websocketConnectionsRejected atomic.Uint32

	websocketConnectionMessagesReceived atomic.Uint32
	websocketConnectionMessagesSent     atomic.Uint32

	websocketConnectionMessagesBytesReceived atomic.Uint64
	websocketConnectionMessagesBytesSent     atomic.Uint64
}

func New(name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, messageHandlers WebsocketMessageHandlers, eventHandler Event.Handler) (*WebsocketServer, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("config.TcpServerConfig is nil")
	}
	if config.Pattern == "" {
		return nil, errors.New("config.Pattern is empty")
	}
	if messageHandlers == nil {
		return nil, errors.New("messageHandlers is nil")
	}
	if config.TopicManagerConfig == nil {
		return nil, errors.New("config.TopicManager is nil")
	}
	if config.Upgrader == nil {
		config.Upgrader = &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}

	server := &WebsocketServer{
		name:       name,
		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		sessionManager: Tools.NewSessionManager(name+"_sessionManager", config.SessionManagerConfig, nil, nil),

		config:            config,
		connectionChannel: make(chan *websocket.Conn),
	}
	topicHandlers := make(Tools.TopicHandlers)
	for topic, handler := range messageHandlers {
		topicHandlers[topic] = server.toTopicHandler(handler)
	}
	server.topicManager = Tools.NewTopicManager(config.TopicManagerConfig, topicHandlers, nil)

	server.httpServer = HTTPServer.New(server.name+"_httpServer",
		&Config.HTTPServer{
			TcpServerConfig: server.config.TcpServerConfig,
		},
		whitelist, blacklist,
		map[string]http.HandlerFunc{
			server.config.Pattern: server.getHTTPWebsocketUpgradeHandler(),
		},
		eventHandler,
	)
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
		Event.Function:      Event.GetCallerFuncName(2),
		Event.InstanceId:    server.instanceId,
		Event.SessionId:     server.sessionId,
	}
}
