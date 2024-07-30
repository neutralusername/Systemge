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

func (node *Node) RetrieveWebsocketBytesSentCounter() uint64 {
	if websocket := node.websocket; websocket != nil {
		return websocket.bytesSentCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveWebsocketBytesReceivedCounter() uint64 {
	if websocket := node.websocket; websocket != nil {
		return websocket.bytesReceivedCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveWebsocketIncomingMessageCounter() uint32 {
	if websocket := node.websocket; websocket != nil {
		return websocket.incomingMessageCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveWebsocketOutgoingMessageCounter() uint32 {
	if websocket := node.websocket; websocket != nil {
		return websocket.outgoigMessageCounter.Swap(0)
	}
	return 0
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
	node.websocket.httpServer = HTTP.New(&Config.HTTP{
		ServerConfig: node.websocket.application.GetWebsocketComponentConfig().ServerConfig,
	}, map[string]http.HandlerFunc{
		node.websocket.application.GetWebsocketComponentConfig().Pattern: websocket.websocketUpgrade(node.GetInternalWarningError()),
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
	return nil
}
