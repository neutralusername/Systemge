package WebsocketServer

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type WebsocketServer struct {
	name string

	status      int
	statusMutex sync.Mutex

	config     *Config.WebsocketServer
	mailer     *Tools.Mailer
	randomizer *Tools.Randomizer

	ipRateLimiter *Tools.IpRateLimiter

	httpServer *HTTPServer.HTTPServer

	connectionChannel chan *websocket.Conn

	clients         map[string]*WebsocketClient            // websocketId -> client
	groups          map[string]map[string]*WebsocketClient // groupId -> map[websocketId]client
	clientGroups    map[string]map[string]bool             // websocketId -> map[groupId]bool
	messageHandlers MessageHandlers
	clientMutex     sync.RWMutex

	stopChannel chan bool
	waitGroup   sync.WaitGroup

	messageHandlerMutex sync.Mutex

	onErrorHandler   func(*Event.Event) *Event.Event
	onWarningHandler func(*Event.Event) *Event.Event
	onInfoHandler    func(*Event.Event) *Event.Event

	// metrics

	acceptedWebsocketConnectionsCounter atomic.Uint32
	rejectedWebsocketConnectionsCounter atomic.Uint32

	incomingMessageCounter atomic.Uint32
	outgoigMessageCounter  atomic.Uint32

	failedSendCounter atomic.Uint32

	bytesSentCounter     atomic.Uint64
	bytesReceivedCounter atomic.Uint64
}

// onConnectHandler, onDisconnectHandler may be nil.
func New(name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, messageHandlers MessageHandlers) *WebsocketServer {
	if config == nil {
		panic("config is nil")
	}
	if config.TcpServerConfig == nil {
		panic("config.ServerConfig is nil")
	}
	if config.Pattern == "" {
		panic("config.Pattern is empty")
	}
	if messageHandlers == nil {
		panic("messageHandlers is nil")
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
		name:            name,
		clients:         make(map[string]*WebsocketClient),
		groups:          make(map[string]map[string]*WebsocketClient),
		clientGroups:    make(map[string]map[string]bool),
		messageHandlers: messageHandlers,
		config:          config,
		mailer:          Tools.NewMailer(config.MailerConfig),
		randomizer:      Tools.NewRandomizer(config.RandomizerSeed),
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
	return server
}

func (server *WebsocketServer) Start() error {
	return server.start(true)
}
func (server *WebsocketServer) start(lock bool) error {
	if lock {
		server.statusMutex.Lock()
		defer server.statusMutex.Unlock()
	}

	if event := server.onInfo(Event.NewInfo(
		Event.StartingService,
		"starting websocketServer",
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
			"websocketServer already started",
			server.GetServerContext().Merge(Event.Context{}),
		))
		return errors.New("failed to start websocketServer")
	}
	server.status = Status.Pending

	if server.config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(server.config.IpRateLimiter)
	}
	server.connectionChannel = make(chan *websocket.Conn)
	if err := server.httpServer.Start(); err != nil { // TODO: context from this service missing - handle this somehow
		if server.ipRateLimiter != nil {
			server.ipRateLimiter.Close()
			server.ipRateLimiter = nil
			close(server.connectionChannel)
		}
		server.status = Status.Stoped
		return err
	}

	server.stopChannel = make(chan bool)
	go server.receiveWebsocketConnectionLoop()

	server.status = Status.Started

	if event := server.onInfo(Event.NewInfo(
		Event.ServiceStarted,
		"websocketServer started",
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
		"stopping websocketServer",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Started {
		server.onWarning(Event.NewWarning(
			Event.ServiceAlreadyStarted,
			"websocketServer not started",
			Event.NoOption,
			Event.NoOption,
			Event.NoOption,
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
	close(server.connectionChannel)

	server.status = Status.Stoped
	event := server.onInfo(Event.NewInfo(
		Event.ServiceStopped,
		"websocketServer stopped",
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
		"service":       Event.WebsocketServer,
		"serviceName":   server.name,
		"serviceStatus": Status.ToString(server.status),
		"function":      Event.GetCallerFuncName(2),
	}
	return ctx
}
