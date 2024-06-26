package Node

import (
	"Systemge/Error"

	"github.com/gorilla/websocket"
)

func (node *Node) startWebsocketComponent() error {
	err := node.startWebsocketHandshakeHTTPServer()
	if err != nil {
		return Error.New("Error starting websocket handshake handler", err)
	}
	node.websocketClients = make(map[string]*WebsocketClient)
	node.websocketConnChannel = make(chan *websocket.Conn, WEBSOCKETCONNCHANNEL_BUFFERSIZE)
	go node.handleWebsocketConnections()
	return nil
}

func (node *Node) stopWebsocketComponent() error {
	err := node.stopWebsocketHandshakeHTTPServer()
	if err != nil {
		return Error.New("Error stopping websocket handshake handler", err)
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
	return nil
}

func (node *Node) startWebsocketHandshakeHTTPServer() error {
	handlers := map[string]HTTPRequestHandler{
		node.config.WebsocketPattern: node.promoteToWebsocket(),
	}
	httpServer := createHTTPServer(node.config.WebsocketPort, handlers)
	err := startHTTPServer(httpServer, node.config.WebsocketCertPath, node.config.WebsocketKeyPath)
	if err != nil {
		return Error.New("Error starting websocket handshake handler", err)
	}
	node.websocketHandshakeHTTPServer = httpServer
	return nil
}

func (node *Node) stopWebsocketHandshakeHTTPServer() error {
	err := stopHTTPServer(node.websocketHandshakeHTTPServer)
	if err != nil {
		return Error.New("Error stopping websocket handshake handler", err)
	}
	node.websocketHandshakeHTTPServer = nil
	return nil
}
