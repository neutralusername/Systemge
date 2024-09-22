package WebsocketServer

import (
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
	httpServer := HTTPServer.New(server.name+"_httpServer",
		&Config.HTTPServer{
			TcpServerConfig: server.config.TcpServerConfig,
		},
		whitelist, blacklist,
		map[string]http.HandlerFunc{
			server.config.Pattern: server.getHTTPWebsocketUpgradeHandler(),
		},
	)
	server.httpServer = httpServer
	return server
}

func (server *WebsocketServer) Start() *Event.Event {
	return server.start(true)
}
func (server *WebsocketServer) start(lock bool) *Event.Event {
	if lock {
		server.statusMutex.Lock()
		defer server.statusMutex.Unlock()
	}

	if event := server.onInfo(Event.New(Event.StartingService,
		server.GetServerContext().Merge(Event.Context{
			"info": "starting websocketServer",
		}),
	)); event.IsError() {
		return event
	}

	if server.status != Status.Stoped {
		return server.onError(Event.New(
			Event.ServiceAlreadyStarted,
			server.GetServerContext().Merge(Event.Context{
				"error": "failed to start websocketServer",
			}),
		))
	}
	server.status = Status.Pending

	if server.config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(server.config.IpRateLimiter)
	}
	server.connectionChannel = make(chan *websocket.Conn)
	if event := server.httpServer.Start(); event.IsError() {
		if server.ipRateLimiter != nil {
			server.ipRateLimiter.Close()
			server.ipRateLimiter = nil
			close(server.connectionChannel)
		}
		server.status = Status.Stoped
		return event // TODO: context from this service missing - handle this somehow
	}

	server.stopChannel = make(chan bool)
	go server.receiveWebsocketConnectionLoop()

	server.status = Status.Started
	event := server.onInfo(Event.New(
		Event.ServiceStarted,
		server.GetServerContext().Merge(Event.Context{
			"info": "websocketServer started",
		}),
	))
	if event.IsError() {
		if event := server.stop(false); event.IsError() {
			panic(event.Marshal())
		}
	}
	return event
}

func (server *WebsocketServer) Stop() *Event.Event {
	return server.stop(true)
}
func (server *WebsocketServer) stop(lock bool) *Event.Event {
	if lock {
		server.statusMutex.Lock()
		defer server.statusMutex.Unlock()
	}

	if event := server.onInfo(Event.New(
		Event.StoppingService,
		server.GetServerContext().Merge(Event.Context{
			"info": "stopping websocketServer",
		}),
	)); event.IsError() {
		return event
	}

	if server.status != Status.Started {
		return server.onError(Event.New(
			Event.ServiceAlreadyStopped,
			server.GetServerContext().Merge(Event.Context{
				"error": "failed to stop websocketServer",
			}),
		))
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
	event := server.onInfo(Event.New(
		Event.ServiceStopped,
		server.GetServerContext().Merge(Event.Context{
			"info": "websocketServer stopped",
		}),
	))
	if event.IsError() {
		if event := server.start(false); event.IsError() {
			panic(event.Marshal())
		}
	}
	return event
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
		"serviceType":   Event.ServiceTypeWebsocketServer,
		"serviceName":   server.name,
		"serviceStatus": Status.ToString(server.status),
		//"caller":        Event.GetCallerPath(2),
	}
	return ctx
}
