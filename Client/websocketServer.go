package Client

import (
	"Systemge/Error"

	"github.com/gorilla/websocket"
)

func (client *Client) startWebsocketServer() error {
	err := client.startWebsocketHandshakeHTTPServer()
	if err != nil {
		return Error.New("Error starting websocket handshake handler", err)
	}
	client.websocketClients = make(map[string]*WebsocketClient)
	client.websocketConnChannel = make(chan *websocket.Conn, WEBSOCKETCONNCHANNEL_BUFFERSIZE)
	go client.handleWebsocketConnections()
	return nil
}

func (client *Client) stopWebsocketServer() error {
	err := client.stopWebsocketHandshakeHTTPServer()
	if err != nil {
		return Error.New("Error stopping websocket handshake handler", err)
	}
	close(client.websocketConnChannel)

	client.websocketMutex.Lock()
	clientsToDisconnect := make([]*WebsocketClient, 0)
	for _, websocketClient := range client.websocketClients {
		clientsToDisconnect = append(clientsToDisconnect, websocketClient)
	}
	client.websocketMutex.Unlock()

	for _, websocketClient := range clientsToDisconnect {
		websocketClient.Disconnect()
	}
	return nil
}

func (client *Client) startWebsocketHandshakeHTTPServer() error {
	handlers := map[string]HTTPRequestHandler{
		client.config.WebsocketPattern: client.promoteToWebsocket(),
	}
	httpServer := createHTTPServer(client.config.WebsocketPort, handlers)
	err := startHTTPServer(httpServer, client.config.WebsocketCertPath, client.config.WebsocketKeyPath)
	if err != nil {
		return Error.New("Error starting websocket handshake handler", err)
	}
	client.websocketHandshakeHTTPServer = httpServer
	return nil
}

func (client *Client) stopWebsocketHandshakeHTTPServer() error {
	err := stopHTTPServer(client.websocketHandshakeHTTPServer)
	if err != nil {
		return Error.New("Error stopping websocket handshake handler", err)
	}
	client.websocketHandshakeHTTPServer = nil
	return nil
}
