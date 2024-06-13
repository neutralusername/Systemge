package Module

import (
	"Systemge/HTTP"
	"Systemge/MessageBrokerClient"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
)

func NewWebsocketClient(name string, messageBrokerClientPort string, loggerPath string, pattern string, websocketPort string, tlsCert string, tlsKey string, newWebsocketApplication NewWebsocketApplicationFunc) *MessageBrokerClient.Client {
	logger := Utilities.NewLogger(loggerPath)
	httpServer := HTTP.New(websocketPort, name+"HTTP", tlsCert, tlsKey, logger)
	websocketServer := WebsocketServer.NewWebsocketServer(name, logger, httpServer)
	httpServer.RegisterPattern(pattern, WebsocketServer.PromoteToWebsocket(websocketServer))
	messageBrokerClient := MessageBrokerClient.New(name, messageBrokerClientPort, logger, websocketServer)
	websocketApplication := newWebsocketApplication(logger, messageBrokerClient)
	messageBrokerClient.SetApplication(websocketApplication)
	websocketServer.SetWebsocketApplication(websocketApplication)
	return messageBrokerClient
}
