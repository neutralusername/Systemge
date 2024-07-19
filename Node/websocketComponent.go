package Node

import (
	"Systemge/Error"
	"Systemge/Http"
)

func (node *Node) startWebsocketComponent() error {
	handlers := map[string]Http.RequestHandler{
		node.GetWebsocketComponent().GetWebsocketComponentConfig().Pattern: node.WebsocketUpgrade(),
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
