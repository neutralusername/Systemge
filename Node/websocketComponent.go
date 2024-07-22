package Node

import (
	"Systemge/Error"
	"Systemge/Http"
	"net/http"
)

func (node *Node) startWebsocketComponent() error {
	handlers := map[string]http.HandlerFunc{
		node.GetWebsocketComponent().GetWebsocketComponentConfig().Pattern: Http.AccessControllWrapper(node.WebsocketUpgrade(), node.websocketBlacklist, node.websocketWhitelist),
	}
	httpServer := Http.New(node.GetWebsocketComponent().GetWebsocketComponentConfig().Http.Server.Port, handlers)
	err := Http.Start(httpServer, node.GetWebsocketComponent().GetWebsocketComponentConfig().Http.Server.TlsCertPath, node.GetWebsocketComponent().GetWebsocketComponentConfig().Http.Server.TlsKeyPath)
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
