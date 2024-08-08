package Node

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

type websocketComponent struct {
	config              *Config.Websocket
	mutex               sync.RWMutex
	httpServer          *HTTP.Server
	connChannel         chan *websocket.Conn
	clients             map[string]*WebsocketClient            // websocketId -> websocketClient
	groups              map[string]map[string]*WebsocketClient // groupId -> map[websocketId]websocketClient
	clientGroups        map[string]map[string]bool             // websocketId -> map[groupId]bool
	onDisconnectHandler func(websocketClient *WebsocketClient)
	onConnectHandler    func(websocketClient *WebsocketClient)
	messageHandlers     map[string]WebsocketMessageHandler
	messageHandlerMutex sync.Mutex
	handleMessage       func(websocketClient *WebsocketClient, message *Message.Message) error

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	nodeName      string

	incomingMessageCounter atomic.Uint32
	outgoigMessageCounter  atomic.Uint32
	bytesSentCounter       atomic.Uint64
	bytesReceivedCounter   atomic.Uint64
}

func (node *Node) startWebsocketComponent() (*websocketComponent, error) {
	if node.newNodeConfig.WebsocketConfig.Upgrader == nil {
		node.newNodeConfig.WebsocketConfig.Upgrader = &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}
	application := node.application.(WebsocketComponent)
	websocket := &websocketComponent{
		connChannel:     make(chan *websocket.Conn),
		clients:         make(map[string]*WebsocketClient),
		groups:          make(map[string]map[string]*WebsocketClient),
		clientGroups:    make(map[string]map[string]bool),
		messageHandlers: map[string]WebsocketMessageHandler{},
		config:          node.newNodeConfig.WebsocketConfig,
		infoLogger:      node.GetInternalInfoLogger(),
		warningLogger:   node.GetInternalWarningLogger(),
		errorLogger:     node.GetErrorLogger(),
		nodeName:        node.GetName(),
	}
	websocket.handleMessage = func(websocketClient *WebsocketClient, message *Message.Message) error {
		websocket.messageHandlerMutex.Lock()
		handler := websocket.messageHandlers[message.GetTopic()]
		websocket.messageHandlerMutex.Unlock()
		if handler == nil {
			err := websocketClient.Send(Message.NewAsync("error", Error.New("no handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", nil).Error()).Serialize())
			if err != nil {
				if warningLogger := websocket.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send error message to websocket client", err).Error())
				}
			}
			return Error.New("no handler for topic \""+message.GetTopic()+"\"", nil)
		}
		err := handler(node, websocketClient, message)
		if err != nil {
			err := websocketClient.Send(Message.NewAsync("error", Error.New("error in handler for topic \""+message.GetTopic()+"\" from websocketClient \""+websocketClient.GetId()+"\"", err).Error()).Serialize())
			if err != nil {
				if warningLogger := websocket.warningLogger; warningLogger != nil {
					warningLogger.Log(Error.New("Failed to send error message to websocket client", err).Error())
				}
			}
		}
		return nil
	}

	websocket.httpServer = HTTP.New(&Config.HTTP{
		ServerConfig: websocket.config.ServerConfig,
	}, map[string]http.HandlerFunc{
		websocket.config.Pattern: websocket.websocketUpgrade(websocket.warningLogger),
	})
	websocket.onDisconnectHandler = func(websocketClient *WebsocketClient) {
		application.OnDisconnectHandler(node, websocketClient)
	}
	websocket.onConnectHandler = func(websocketClient *WebsocketClient) {
		application.OnConnectHandler(node, websocketClient)
	}
	websocket.messageHandlers = application.GetWebsocketMessageHandlers()
	err := websocket.httpServer.Start()
	if err != nil {
		return nil, Error.New("failed starting websocket handshake handler", err)
	}
	go websocket.handleWebsocketConnections()
	return websocket, nil
}

func (node *Node) stopWebsocketComponent() {
	websocket := node.websocket
	node.websocket = nil
	websocket.httpServer.Stop()
	websocket.httpServer = nil
	close(websocket.connChannel)
	websocket.mutex.Lock()
	websocketClientsToDisconnect := make([]*WebsocketClient, 0)
	for _, websocketClient := range websocket.clients {
		websocketClientsToDisconnect = append(websocketClientsToDisconnect, websocketClient)
	}
	websocket.mutex.Unlock()
	for _, websocketClient := range websocketClientsToDisconnect {
		websocketClient.Disconnect()
	}
}
