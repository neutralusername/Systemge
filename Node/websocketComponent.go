package Node

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Http"
	"net/http"
)

func (node *Node) startWebsocketComponent() error {
	httpServer := Http.New(&Config.Http{
		Server: node.GetWebsocketComponent().GetWebsocketComponentConfig().Server,
		Handlers: map[string]http.HandlerFunc{
			node.GetWebsocketComponent().GetWebsocketComponentConfig().Pattern: node.WebsocketUpgrade(),
		},
	})
	err := httpServer.Start()
	if err != nil {
		return Error.New("failed starting websocket handshake handler", err)
	}
	node.websocketHttpServer = httpServer
	node.websocketClients = make(map[string]*WebsocketClient)
	go node.handleWebsocketConnections()
	node.websocketStarted = true
	return nil
}

func (node *Node) stopWebsocketComponent() error {
	err := node.websocketHttpServer.Stop()
	if err != nil {
		return Error.New("failed stopping websocket handshake handler", err)
	}
	node.websocketHttpServer = nil
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
