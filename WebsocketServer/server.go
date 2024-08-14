package WebsocketServer

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type WebsocketMessageHandler func(*Client, *Message.Message) error

type WebsocketServer struct {
	config        *Config.WebsocketServer
	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
	randomizer    *Tools.Randomizer

	onConnectHandler    func(websocketClient *Client)
	onDisconnectHandler func(websocketClient *Client)

	httpServer *HTTPServer.Server

	connChannel     chan *websocket.Conn
	clients         map[string]*Client            // websocketId -> websocketClient
	groups          map[string]map[string]*Client // groupId -> map[websocketId]websocketClient
	clientGroups    map[string]map[string]bool    // websocketId -> map[groupId]bool
	messageHandlers map[string]WebsocketMessageHandler

	messageHandlerMutex sync.Mutex
	mutex               sync.RWMutex

	// metrics

	incomingMessageCounter atomic.Uint32
	outgoigMessageCounter  atomic.Uint32
	bytesSentCounter       atomic.Uint64
	bytesReceivedCounter   atomic.Uint64
}

func New(config *Config.WebsocketServer, messageHandlers map[string]WebsocketMessageHandler, onConnectHandler func(*Client), onDisconnectHandler func(*Client)) *WebsocketServer {
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
		connChannel:         make(chan *websocket.Conn),
		clients:             make(map[string]*Client),
		groups:              make(map[string]map[string]*Client),
		clientGroups:        make(map[string]map[string]bool),
		messageHandlers:     messageHandlers,
		onConnectHandler:    onConnectHandler,
		onDisconnectHandler: onDisconnectHandler,
		config:              config,
		errorLogger:         Tools.NewLogger("[Error: \"WebsocketServer\"] ", config.ErrorLoggerPath),
		infoLogger:          Tools.NewLogger("[Info: \"WebsocketServer\"] ", config.InfoLoggerPath),
		warningLogger:       Tools.NewLogger("[Warning: \"WebsocketServer\" (internal)] ", config.WarningLoggerPath),
		mailer:              Tools.NewMailer(config.Mailer),
		randomizer:          Tools.NewRandomizer(config.RandomizerSeed),
	}
	server.httpServer = HTTPServer.New(&Config.HTTP{
		ServerConfig: server.config.ServerConfig,
	}, map[string]http.HandlerFunc{
		server.config.Pattern: server.websocketUpgrade(server.warningLogger),
	})
	return server
}

func (server *WebsocketServer) Start() error {
	err := server.httpServer.Start()
	if err != nil {
		return Error.New("failed starting websocket handshake handler", err)
	}
	server.handleWebsocketConnections()
	return nil
}

func (server *WebsocketServer) Stop() error {
	server.httpServer.Stop()
	server.httpServer = nil
	close(server.connChannel)
	server.mutex.Lock()
	websocketClientsToDisconnect := make([]*Client, 0)
	for _, websocketClient := range server.clients {
		websocketClientsToDisconnect = append(websocketClientsToDisconnect, websocketClient)
	}
	server.mutex.Unlock()
	for _, websocketClient := range websocketClientsToDisconnect {
		websocketClient.Disconnect()
	}
	return nil
}

func (server *WebsocketServer) AddMessageHandler(topic string, handler WebsocketMessageHandler) {
	server.messageHandlerMutex.Lock()
	server.messageHandlers[topic] = handler
	server.messageHandlerMutex.Unlock()
}

func (server *WebsocketServer) RemoveMessageHandler(topic string) {
	server.messageHandlerMutex.Lock()
	delete(server.messageHandlers, topic)
	server.messageHandlerMutex.Unlock()
}
