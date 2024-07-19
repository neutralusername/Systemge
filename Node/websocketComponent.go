package Node

import (
	"Systemge/Error"
	"Systemge/Http"
	"net/http"

	"github.com/gorilla/websocket"
)

func (node *Node) startWebsocketComponent() error {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(httpRequest *http.Request) bool {
			return true
		},
	}
	handlers := map[string]Http.RequestHandler{
		node.GetWebsocketComponent().GetWebsocketComponentConfig().Pattern: Http.WebsocketUpgrade(upgrader, node.GetWarningLogger(), &node.isStarted, node.websocketConnChannel),
	}
	httpServer := Http.New(node.GetWebsocketComponent().GetWebsocketComponentConfig().Server.Port, handlers)
	err := Http.Start(httpServer, node.GetWebsocketComponent().GetWebsocketComponentConfig().Server.TlsCertPath, node.GetWebsocketComponent().GetWebsocketComponentConfig().Server.TlsKeyPath)
	if err != nil {
		return Error.New("failed starting websocket handshake handler", err)
	}
	node.websocketHandshakeHTTPServer = httpServer
	node.websocketClients = make(map[string]*WebsocketClient)
	go node.handleWebsocketConnections()
	node.websocketStarted = true
	return nil
}

func (node *Node) stopWebsocketComponent() error {
	err := Http.Stop(node.websocketHandshakeHTTPServer)
	if err != nil {
		return Error.New("failed stopping websocket handshake handler", err)
	}
	node.websocketHandshakeHTTPServer = nil

	node.websocketMutex.Lock()
	websocketClientsToDisconnect := make([]*WebsocketClient, 0)
	for _, websocketClient := range node.websocketClients {
		websocketClientsToDisconnect = append(websocketClientsToDisconnect, websocketClient)
	}
	node.websocketMutex.Unlock()

	for _, websocketClient := range websocketClientsToDisconnect {
		websocketClient.Disconnect()
	}
	node.websocketStarted = false
	return nil
}
