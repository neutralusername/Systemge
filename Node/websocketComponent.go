package Node

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTP"

	"github.com/gorilla/websocket"
)

type websocketComponent struct {
	config              *Config.Websocket
	application         WebsocketComponent
	mutex               sync.Mutex
	httpServer          *HTTP.Server
	connChannel         chan *websocket.Conn
	clients             map[string]*WebsocketClient            // websocketId -> websocketClient
	groups              map[string]map[string]*WebsocketClient // groupId -> map[websocketId]websocketClient
	clientGroups        map[string]map[string]bool             // websocketId -> map[groupId]bool
	onDisconnectWraper  func(websocketClient *WebsocketClient)
	messageHandlerMutex sync.Mutex

	incomingMessageCounter atomic.Uint32
	outgoigMessageCounter  atomic.Uint32
	bytesSentCounter       atomic.Uint64
	bytesReceivedCounter   atomic.Uint64
}

func (node *Node) startWebsocketComponent() error {
	websocketComponent := &websocketComponent{
		application:  node.application.(WebsocketComponent),
		connChannel:  make(chan *websocket.Conn),
		clients:      make(map[string]*WebsocketClient),
		groups:       make(map[string]map[string]*WebsocketClient),
		clientGroups: make(map[string]map[string]bool),
		config:       node.newNodeConfig.WebsocketConfig,
	}
	if websocketComponent.config.Upgrader == nil {
		websocketComponent.config.Upgrader = &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}
	node.websocket = websocketComponent
	node.websocket.httpServer = HTTP.New(&Config.HTTP{
		ServerConfig: node.websocket.config.ServerConfig,
	}, map[string]http.HandlerFunc{
		node.websocket.config.Pattern: websocketComponent.websocketUpgrade(node.GetInternalWarningError()),
	})
	node.websocket.onDisconnectWraper = func(websocketClient *WebsocketClient) {
		websocketComponent.application.OnDisconnectHandler(node, websocketClient)
	}
	err := node.websocket.httpServer.Start()
	if err != nil {
		return Error.New("failed starting websocket handshake handler", err)
	}
	go node.handleWebsocketConnections()
	return nil
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
