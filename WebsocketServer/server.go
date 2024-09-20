package WebsocketServer

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Service"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type WebsocketServer struct {
	name string

	status      int
	statusMutex sync.Mutex

	config     *Config.WebsocketServer
	mailer     *Tools.Mailer
	randomizer *Tools.Randomizer

	onConnectHandler    OnConnectHandler
	onDisconnectHandler OnDisconnectHandler

	ipRateLimiter *Tools.IpRateLimiter

	httpServer *HTTPServer.HTTPServer

	connectionChannel chan *websocket.Conn
	clients           map[string]*WebsocketClient            // websocketId -> client
	groups            map[string]map[string]*WebsocketClient // groupId -> map[websocketId]client
	clientGroups      map[string]map[string]bool             // websocketId -> map[groupId]bool
	messageHandlers   MessageHandlers
	clientMutex       sync.RWMutex

	messageHandlerMutex sync.Mutex

	onErrorHandler   func(*Event.Event) *Event.Event
	onWarningHandler func(*Event.Event) *Event.Event
	onInfoHandler    func(*Event.Event) *Event.Event

	// metrics

	incomingMessageCounter atomic.Uint32
	outgoigMessageCounter  atomic.Uint32
	bytesSentCounter       atomic.Uint64
	bytesReceivedCounter   atomic.Uint64
}

// onConnectHandler, onDisconnectHandler may be nil.
func New(name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, messageHandlers MessageHandlers, onConnectHandler func(*WebsocketClient) error, onDisconnectHandler func(*WebsocketClient)) *WebsocketServer {
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
		name:                name,
		clients:             make(map[string]*WebsocketClient),
		groups:              make(map[string]map[string]*WebsocketClient),
		clientGroups:        make(map[string]map[string]bool),
		messageHandlers:     messageHandlers,
		onConnectHandler:    onConnectHandler,
		onDisconnectHandler: onDisconnectHandler,
		config:              config,
		mailer:              Tools.NewMailer(config.MailerConfig),
		randomizer:          Tools.NewRandomizer(config.RandomizerSeed),
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
	if server.config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(server.config.IpRateLimiter)
	}
	server.httpServer = httpServer
	return server
}

func (server *WebsocketServer) GetServerContext(context Event.Context) Event.Context {
	ctx := Event.Context{
		"serviceType": Service.WebsocketServer,
		"name":        server.name,
		"status":      Status.ToString(server.status),
		"caller":      Event.GetCallerPath(2),
	}
	for key, value := range context {
		ctx[key] = value
	}
	return ctx
}

func (server *WebsocketServer) Start() *Event.Event {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.Stoped {
		return server.onError(Event.New(
			Event.AlreadyStarted,
			server.GetServerContext(
				Event.Context{
					"error": "failed to start websocketServer",
				},
			),
		))
	}

	server.onInfo(Event.New(Event.StartingService, server.GetServerContext(nil)))
	server.status = Status.Pending

	server.connectionChannel = make(chan *websocket.Conn)
	err := server.httpServer.Start()
	if err != nil {
		server.ipRateLimiter.Close()
		server.ipRateLimiter = nil
		close(server.connectionChannel)
		server.connectionChannel = nil
		server.status = Status.Stoped
		return server.onError(Event.New(
			Event.FailedStartingService,
			server.GetServerContext(Event.Context{
				"error":             err.Error(),
				"targetServiceType": Service.HttpServer,
				"targetServiceName": server.httpServer.GetName(),
			}),
		))
	}
	go server.handleWebsocketConnections()

	server.status = Status.Started
	return server.onInfo(Event.New(Event.ServiceStarted, server.GetServerContext(nil)))
}

func (server *WebsocketServer) Stop() *Event.Event {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.Started {
		return server.onError(Event.New(
			Event.AlreadyStopped,
			server.GetServerContext(Event.Context{
				"error": "failed to stop websocketServer",
			}),
		))
	}

	server.onInfo(Event.New(Event.StoppingService, server.GetServerContext(nil)))
	server.status = Status.Pending

	server.httpServer.Stop()
	close(server.connectionChannel)
	if server.ipRateLimiter != nil {
		server.ipRateLimiter.Close()
		server.ipRateLimiter = nil
	}

	server.clientMutex.Lock()
	websocketClientsToDisconnect := make([]*WebsocketClient, 0)
	for _, websocketClient := range server.clients {
		websocketClientsToDisconnect = append(websocketClientsToDisconnect, websocketClient)
	}
	server.clientMutex.Unlock()

	for _, websocketClient := range websocketClientsToDisconnect {
		websocketClient.Disconnect()
	}

	server.status = Status.Stoped
	return server.onInfo(Event.New(Event.ServiceStopped, server.GetServerContext(nil)))
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
