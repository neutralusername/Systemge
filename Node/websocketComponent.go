package Node

import (
	"Systemge/Error"

	"github.com/gorilla/websocket"
)

func (node *Node) startWebsocketComponent() error {
	err := node.startWebsocketHandshakeHTTPServer()
	if err != nil {
		return Error.New("failed starting websocket handshake handler", err)
	}
	node.websocketClients = make(map[string]*WebsocketClient)
	node.websocketConnChannel = make(chan *websocket.Conn, WEBSOCKETCONNCHANNEL_BUFFERSIZE)
	go node.handleWebsocketConnections()
	node.websocketStarted = true
	node.config.Logger.Info(Error.New("Started websocket component on node \""+node.GetName()+"\"", nil).Error())
	return nil
}

func (node *Node) stopWebsocketComponent() error {
	err := node.stopWebsocketHandshakeHTTPServer()
	if err != nil {
		return Error.New("failed stopping websocket handshake handler", err)
	}
	close(node.websocketConnChannel)

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
	node.config.Logger.Info(Error.New("Stopped websocket component on node \""+node.GetName()+"\"", nil).Error())
	return nil
}

func (node *Node) startWebsocketHandshakeHTTPServer() error {
	handlers := map[string]HTTPRequestHandler{
		node.websocketComponent.GetWebsocketComponentConfig().Pattern: node.promoteToWebsocket(),
	}
	httpServer := createHTTPServer(node.websocketComponent.GetWebsocketComponentConfig().Server.GetPort(), handlers)
	err := startHTTPServer(httpServer, node.websocketComponent.GetWebsocketComponentConfig().Server.GetTlsCertPath(), node.websocketComponent.GetWebsocketComponentConfig().Server.GetTlsKeyPath())
	if err != nil {
		return Error.New("failed starting websocket handshake handler", err)
	}
	node.websocketHandshakeHTTPServer = httpServer
	return nil
}

func (node *Node) stopWebsocketHandshakeHTTPServer() error {
	err := stopHTTPServer(node.websocketHandshakeHTTPServer)
	if err != nil {
		return Error.New("failed stopping websocket handshake handler", err)
	}
	node.websocketHandshakeHTTPServer = nil
	return nil
}
