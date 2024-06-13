package Application

import (
	"Systemge/Message"
	"Systemge/WebsocketClient"
)

type WebsocketApplication interface {
	Application
	GetWebsocketMessageHandlers() map[string]WebsocketMessageHandler
	OnConnectHandler(*WebsocketClient.Client)
	OnDisconnectHandler(*WebsocketClient.Client)
}
type WebsocketMessageHandler func(*WebsocketClient.Client, *Message.Message) error
