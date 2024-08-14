package WebsocketServer

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTP"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type WebsocketMessageHandler func(*Client, *Message.Message) error

type Server struct {
	config              *Config.WebsocketServer
	mutex               sync.RWMutex
	httpServer          *HTTP.Server
	connChannel         chan *websocket.Conn
	clients             map[string]*Client            // websocketId -> websocketClient
	groups              map[string]map[string]*Client // groupId -> map[websocketId]websocketClient
	clientGroups        map[string]map[string]bool    // websocketId -> map[groupId]bool
	messageHandlers     map[string]WebsocketMessageHandler
	messageHandlerMutex sync.Mutex
	handleMessage       func(websocketClient *Client, message *Message.Message) error

	onConnectHandler    func(websocketClient *Client)
	onDisconnectHandler func(websocketClient *Client)

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	incomingMessageCounter atomic.Uint32
	outgoigMessageCounter  atomic.Uint32
	bytesSentCounter       atomic.Uint64
	bytesReceivedCounter   atomic.Uint64
}

func New(config *Config.WebsocketServer, messageHandlers map[string]WebsocketMessageHandler, onConnectHandler func(*Client), onDisconnectHandler func(*Client)) *Server {
	if config.Upgrader == nil {
		config.Upgrader = &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}
	server := &Server{
		connChannel:     make(chan *websocket.Conn),
		clients:         make(map[string]*Client),
		groups:          make(map[string]map[string]*Client),
		clientGroups:    make(map[string]map[string]bool),
		messageHandlers: messageHandlers,
		config:          config,
		errorLogger:     Tools.NewLogger("[Error: \"WebsocketServer\"] ", config.ErrorLoggerPath),
		infoLogger:      Tools.NewLogger("[Info: \"WebsocketServer\"] ", config.InfoLoggerPath),
		warningLogger:   Tools.NewLogger("[Warning: \"WebsocketServer\" (internal)] ", config.WarningLoggerPath),
		mailer:          Tools.NewMailer(config.Mailer),
	}
	server.handleMessage = func(websocketClient *Client, message *Message.Message) error {
		server.messageHandlerMutex.Lock()
		handler := server.messageHandlers[message.GetTopic()]
		server.messageHandlerMutex.Unlock()
		if handler == nil {
			err := websocketClient.Send(Message.NewAsync("error", Error.New("no handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send error message to websocket client", err).Error())
				}
			}
			return Error.New("no handler for topic \""+message.GetTopic()+"\"", nil)
		}
		err := handler(websocketClient, message)
		if err != nil {
			err := websocketClient.Send(Message.NewAsync("error", Error.New("error in handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", err).Error()).Serialize())
			if err != nil {
				if warningLogger := server.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send error message to websocket client", err).Error())
				}
			}
		}
		return nil
	}
	server.httpServer = HTTP.New(&Config.HTTP{
		ServerConfig: server.config.ServerConfig,
	}, map[string]http.HandlerFunc{
		server.config.Pattern: server.websocketUpgrade(server.warningLogger),
	})
	return server
}

func (server *Server) Start() error {
	err := server.httpServer.Start()
	if err != nil {
		return Error.New("failed starting websocket handshake handler", err)
	}
	server.handleWebsocketConnections()
	return nil
}

func (server *Server) Stop() {
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
}

func (server *Server) AddMessageHandler(topic string, handler WebsocketMessageHandler) {
	server.messageHandlerMutex.Lock()
	server.messageHandlers[topic] = handler
	server.messageHandlerMutex.Unlock()
}

func (server *Server) RemoveMessageHandler(topic string) {
	server.messageHandlerMutex.Lock()
	delete(server.messageHandlers, topic)
	server.messageHandlerMutex.Unlock()
}
