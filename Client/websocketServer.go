package Client

import (
	"Systemge/Utilities"

	"github.com/gorilla/websocket"
)

func (client *Client) StartWebsocketServer() error {
	err := client.StartWebsocketHandshakeHTTPServer()
	if err != nil {
		return Utilities.NewError("Error starting websocket handshake handler", err)
	}
	client.websocketClients = make(map[string]*WebsocketClient)
	client.websocketConnChannel = make(chan *websocket.Conn, WEBSOCKETCONNCHANNEL_BUFFERSIZE)
	go client.handleWebsocketConnections()
	return nil
}

func (client *Client) StopWebsocketServer() error {
	err := client.StopWebsocketHandshakeHTTPServer()
	if err != nil {
		return Utilities.NewError("Error stopping websocket handshake handler", err)
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

func (client *Client) StartWebsocketHandshakeHTTPServer() error {
	handlers := map[string]HTTPRequestHandler{
		client.config.WebsocketPattern: client.PromoteToWebsocket(),
	}
	httpServer := CreateHTTPServer(client.config.WebsocketPort, client.config.WebsocketCertPath, client.config.WebsocketKeyPath, handlers)
	err := StartHTTPServer(httpServer, client.config.WebsocketCertPath, client.config.WebsocketKeyPath)
	if err != nil {
		return Utilities.NewError("Error starting websocket handshake handler", err)
	}
	client.websocketHandshakeHTTPServer = httpServer
	return nil
}

func (client *Client) StopWebsocketHandshakeHTTPServer() error {
	err := StopHTTPServer(client.websocketHandshakeHTTPServer)
	if err != nil {
		return Utilities.NewError("Error stopping websocket handshake handler", err)
	}
	client.websocketHandshakeHTTPServer = nil
	return nil
}
