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

	config     *Config.WebsocketServer
	mailer     *Tools.Mailer
	randomizer *Tools.Randomizer

	httpServer        *HTTPServer.HTTPServer
	connectionChannel chan *websocket.Conn
	ipRateLimiter     *Tools.IpRateLimiter

	websocketConnections       map[string]*WebsocketConnection            // websocketId -> websocketConnection
	websocketConnectionGroups  map[string]map[string]bool                 // websocketId -> map[groupId]bool
	groupsWebsocketConnections map[string]map[string]*WebsocketConnection // groupId -> map[websocketId]websocketConnection
	websocketConnectionMutex   sync.RWMutex

	messageHandlers     MessageHandlers
	messageHandlerMutex sync.Mutex

	onErrorHandler   func(*Event.Event) *Event.Event
	onWarningHandler func(*Event.Event) *Event.Event
	onInfoHandler    func(*Event.Event) *Event.Event

	// metrics

	websocketConnectionsAccepted atomic.Uint32
	websocketConnectionsFailed   atomic.Uint32
	websocketConnectionsRejected atomic.Uint32

	websocketConnectionMessagesReceived atomic.Uint32
	websocketConnectionMessagesSent     atomic.Uint32
	websocketConnectionMessagesFailed   atomic.Uint32

	websocketConnectionMessagesBytesReceived atomic.Uint64
	websocketConnectionMessagesBytesSent     atomic.Uint64
}

// onConnectHandler, onDisconnectHandler may be nil.
func New(name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, messageHandlers MessageHandlers) (*WebsocketServer, error) {
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
		name:                       name,
		websocketConnections:       make(map[string]*WebsocketConnection),
		groupsWebsocketConnections: make(map[string]map[string]*WebsocketConnection),
		websocketConnectionGroups:  make(map[string]map[string]bool),
		instanceId:                 Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		messageHandlers:            messageHandlers,
		config:                     config,
		mailer:                     Tools.NewMailer(config.MailerConfig),
		randomizer:                 Tools.NewRandomizer(config.RandomizerSeed),
		connectionChannel:          make(chan *websocket.Conn),
	}
	server.httpServer = HTTPServer.New(server.name+"_httpServer",
		&Config.HTTPServer{
			TcpServerConfig: server.config.TcpServerConfig,
		},
		whitelist, blacklist,
		map[string]http.HandlerFunc{
			server.config.Pattern: server.getHTTPWebsocketUpgradeHandler(),
		},
	)
	return server, nil
}

func (server *WebsocketServer) Start() error {
	return server.start(true)
}
func (server *WebsocketServer) start(lock bool) error {
	if lock {
		server.statusMutex.Lock()
		defer server.statusMutex.Unlock()
	}

	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)

	if event := server.onInfo(Event.NewInfo(
		Event.StartingService,
		"service websocketServer starting",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Stoped {
		server.onWarning(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"service websocketServer already started",
			server.GetServerContext().Merge(Event.Context{}),
		))
		return errors.New("failed to start websocketServer")
	}
	server.status = Status.Pending

	if server.config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(server.config.IpRateLimiter)
	}
	if err := server.httpServer.Start(); err != nil { // TODO: context from this service missing - handle this somehow
		if server.ipRateLimiter != nil {
			server.ipRateLimiter.Close()
			server.ipRateLimiter = nil
		}
		server.status = Status.Stoped
		return err
	}

	server.stopChannel = make(chan bool)

	server.waitGroup.Add(1)
	go server.receiveWebsocketConnectionLoop()

	server.status = Status.Started

	if event := server.onInfo(Event.NewInfo(
		Event.ServiceStarted,
		"service websocketServer started",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{}),
	)); !event.IsInfo() {
		if err := server.stop(false); err != nil {
			panic(err)
		}
	}
	return nil
}

func (server *WebsocketServer) Stop() error {
	return server.stop(true)
}
func (server *WebsocketServer) stop(lock bool) error {
	if lock {
		server.statusMutex.Lock()
		defer server.statusMutex.Unlock()
	}

	if event := server.onInfo(Event.NewInfo(
		Event.StoppingService,
		"service websocketServer stopping",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status == Status.Stoped {
		server.onWarning(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"service websocketServer already stopped",
			server.GetServerContext().Merge(Event.Context{}),
		))
		return errors.New("websocketServer not started")
	}
	server.status = Status.Pending

	server.httpServer.Stop()
	if server.ipRateLimiter != nil {
		server.ipRateLimiter.Close()
		server.ipRateLimiter = nil
	}

	close(server.stopChannel)
	server.waitGroup.Wait()

	server.status = Status.Stoped
	event := server.onInfo(Event.NewInfo(
		Event.ServiceStopped,
		"service websocketServer stopped",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{}),
	))
	if !event.IsInfo() {
		if err := server.start(false); err != nil {
			panic(err)
		}
	}
	return nil
}

func (server *WebsocketServer) GetName() string {
	return server.name
}

func (server *WebsocketServer) GetStatus() int {
	return server.status
}

func (server *WebsocketServer) AddMessageHandler(topic string, handler MessageHandler) {
	server.messageHandlerMutex.Lock()
	server.messageHandlers[topic] = handler
	server.messageHandlerMutex.Unlock()
}

func (server *WebsocketServer) RemoveMessageHandler(topic string) {
	server.messageHandlerMutex.Lock()
	delete(server.messageHandlers, topic)
	server.messageHandlerMutex.Unlock()
}

func (server *WebsocketServer) onError(event *Event.Event) *Event.Event {
	if server.onErrorHandler != nil {
		return server.onErrorHandler(event)
	}
	return event
}

func (server *WebsocketServer) onWarning(event *Event.Event) *Event.Event {
	if server.onWarningHandler != nil {
		return server.onWarningHandler(event)
	}
	return event
}

func (server *WebsocketServer) onInfo(event *Event.Event) *Event.Event {
	if server.onInfoHandler != nil {
		return server.onInfoHandler(event)
	}
	return event
}

func (server *WebsocketServer) GetServerContext() Event.Context {
	ctx := Event.Context{
		Event.Service:       Event.WebsocketServer,
		Event.ServiceName:   server.name,
		Event.ServiceStatus: Status.ToString(server.status),
		Event.Function:      Event.GetCallerFuncName(2),
		Event.InstanceId:    server.instanceId,
		Event.SessionId:     server.sessionId,
	}
	return ctx
}
