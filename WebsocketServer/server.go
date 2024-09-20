package WebsocketServer

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/HTTPServer"
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

func (server *WebsocketServer) GetServerContext() []*Event.Context {
	return []*Event.Context{
		Event.NewContext("service", "WebsocketServer"),
		Event.NewContext("name", server.name),
	}
}

func (server *WebsocketServer) Start() *Event.Event {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STOPPED {
		return server.onErrorHandler(Event.New(Error.ErrAlreadyStarted, server.GetServerContext()...))
	}

	server.onInfoHandler(Event.New(Event.EventStarting, server.GetServerContext()...))
	server.status = Status.PENDING

	server.connectionChannel = make(chan *websocket.Conn)
	err := server.httpServer.Start()
	if err != nil {
		server.ipRateLimiter.Close()
		server.ipRateLimiter = nil
		close(server.connectionChannel)
		server.connectionChannel = nil
		server.status = Status.STOPPED
		return Event.New("failed starting websocket handshake handler", err)
	}
	go server.handleWebsocketConnections()

	if server.infoLogger != nil {
		server.infoLogger.Log("WebsocketServer started")
	}
	server.status = Status.STARTED
	return nil
}

func (server *WebsocketServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STARTED {
		return Event.New("WebsocketServer is not in started state", nil)
	}
	if server.infoLogger != nil {
		server.infoLogger.Log("Stopping WebsocketServer")
	}
	server.status = Status.PENDING

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

	if server.infoLogger != nil {
		server.infoLogger.Log("WebsocketServer stopped")
	}
	server.status = Status.STOPPED
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

func (server *WebsocketServer) OnError(err error, context string) *Event.Event {
	if server.onErrorHandler != nil {
		return server.onErrorHandler(Event.New(context, err))
	}
	return Event.New(context, err)
}

func (server *WebsocketServer) OnWarning(err error, context string) *Event.Event {
	if server.onWarningHandler != nil {
		return server.onWarningHandler(Event.New(context, err))
	}
	return Event.New(context, err)
}

func (server *WebsocketServer) OnInfo(err error, context string) *Event.Event {
	if server.onErrorHandler != nil {
		return server.onErrorHandler(Event.New(context, err))
	}
	return server.onErrorHandler(Event.New(context, err))
}
