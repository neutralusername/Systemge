package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Http"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type websocketComponent struct {
	application         WebsocketComponent
	mutex               sync.Mutex
	httpServer          *Http.Server
	connChannel         chan *websocket.Conn
	clients             map[string]*WebsocketClient            // websocketId -> websocketClient
	groups              map[string]map[string]*WebsocketClient // groupId -> map[websocketId]websocketClient
	clientGroups        map[string]map[string]bool             // websocketId -> map[groupId]bool
	onDisconnectWraper  func(websocketClient *WebsocketClient)
	messageHandlerMutex sync.Mutex
}

func (node *Node) startWebsocketComponent() error {
	websocket := &websocketComponent{
		application:  node.application.(WebsocketComponent),
		connChannel:  make(chan *websocket.Conn),
		clients:      make(map[string]*WebsocketClient),
		groups:       make(map[string]map[string]*WebsocketClient),
		clientGroups: make(map[string]map[string]bool),
	}
	node.websocket = websocket
	node.websocket.httpServer = Http.New(&Config.Http{
		Server: node.websocket.application.GetWebsocketComponentConfig().Server,
		Handlers: map[string]http.HandlerFunc{
			node.websocket.application.GetWebsocketComponentConfig().Pattern: websocket.websocketUpgrade(node.GetWarningLogger()),
		},
	})
	node.websocket.onDisconnectWraper = func(websocketClient *WebsocketClient) {
		websocket.application.OnDisconnectHandler(node, websocketClient)
	}
	err := node.websocket.httpServer.Start()
	if err != nil {
		return Error.New("failed starting websocket handshake handler", err)
	}
	go node.handleWebsocketConnections()
	return nil
}

func (node *Node) stopWebsocketComponent() error {
	websocket := node.websocket
	node.websocket = nil
	err := websocket.httpServer.Stop()
	if err != nil {
		return Error.New("failed stopping websocket handshake handler", err)
	}
	websocket.httpServer = nil
	websocket.mutex.Lock()
	websocketClientsToDisconnect := make([]*WebsocketClient, 0)
	for _, websocketClient := range websocket.clients {
		websocketClientsToDisconnect = append(websocketClientsToDisconnect, websocketClient)
	}
	websocket.mutex.Unlock()
	for _, websocketClient := range websocketClientsToDisconnect {
		websocketClient.Disconnect()
	}
	return nil
}
