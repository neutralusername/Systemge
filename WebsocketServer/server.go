package WebsocketServer

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type WebsocketServer struct {
	name string

	status      int
	statusMutex sync.Mutex

	config        *Config.WebsocketServer
	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
	randomizer    *Tools.Randomizer

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

	// metrics

	incomingMessageCounter atomic.Uint32
	outgoigMessageCounter  atomic.Uint32
	bytesSentCounter       atomic.Uint64
	bytesReceivedCounter   atomic.Uint64
}

// onConnectHandler, onDisconnectHandler may be nil.
func New(name string, config *Config.WebsocketServer, messageHandlers MessageHandlers, onConnectHandler func(*WebsocketClient) error, onDisconnectHandler func(*WebsocketClient)) *WebsocketServer {
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
		errorLogger:         Tools.NewLogger("[Error: \""+name+"\"] ", config.ErrorLoggerPath),
		infoLogger:          Tools.NewLogger("[Info: \""+name+"\"] ", config.InfoLoggerPath),
		warningLogger:       Tools.NewLogger("[Warning: \""+name+"\"] ", config.WarningLoggerPath),
		mailer:              Tools.NewMailer(config.MailerConfig),
		randomizer:          Tools.NewRandomizer(config.RandomizerSeed),
	}
	return server
}

func (server *WebsocketServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STOPPED {
		return Error.New("WebsocketServer is not in stopped state", nil)
	}
	if server.infoLogger != nil {
		server.infoLogger.Log("Starting WebsocketServer")
	}
	server.status = Status.PENDING

	httpServer := HTTPServer.New(server.name+"_httpServer", &Config.HTTPServer{
		TcpServerConfig: server.config.TcpServerConfig,
	}, map[string]http.HandlerFunc{
		server.config.Pattern: server.getHTTPWebsocketUpgradeHandler(),
	})
	if server.config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(server.config.IpRateLimiter)
	}
	server.connectionChannel = make(chan *websocket.Conn)
	err := httpServer.Start()
	if err != nil {
		server.ipRateLimiter.Close()
		server.ipRateLimiter = nil
		close(server.connectionChannel)
		server.connectionChannel = nil
		server.status = Status.STOPPED
		return Error.New("failed starting websocket handshake handler", err)
	}
	server.httpServer = httpServer
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
		return Error.New("WebsocketServer is not in started state", nil)
	}
	if server.infoLogger != nil {
		server.infoLogger.Log("Stopping WebsocketServer")
	}
	server.status = Status.PENDING

	server.httpServer.Stop()
	server.httpServer = nil
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

func (server *WebsocketServer) GetWhitelist() *Tools.AccessControlList {
	return server.httpServer.GetWhitelist()
}

func (server *WebsocketServer) GetBlacklist() *Tools.AccessControlList {
	return server.httpServer.GetBlacklist()
}
