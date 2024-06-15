package Module

import (
	"Systemge/Client"
	"Systemge/HTTP"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
)

func NewWebsocketClient(name string, clientPort string, loggerPath string, pattern string, websocketPort string, tlsCert string, tlsKey string, newWebsocketApplication NewWebsocketApplicationFunc, args []string) *Client.Client {
	logger := Utilities.NewLogger(loggerPath)
	httpServer := HTTP.New(websocketPort, name+"HTTP", tlsCert, tlsKey, logger)
	websocketServer := WebsocketServer.NewWebsocketServer(name, logger, httpServer)
	httpServer.RegisterPattern(pattern, WebsocketServer.PromoteToWebsocket(websocketServer))
	client := Client.New(name, clientPort, logger, websocketServer)
	websocketApplication := newWebsocketApplication(client, args)
	client.SetApplication(websocketApplication)
	websocketServer.SetWebsocketApplication(websocketApplication)
	return client
}
