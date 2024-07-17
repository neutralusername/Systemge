package Node

import (
	"Systemge/Config"
	"Systemge/Http"
	"Systemge/Message"
)

type Application interface {
}

func ImplementsSystemgeComponent(app Application) bool {
	_, ok := app.(SystemgeComponent)
	return ok
}
func ImplementsHTTPComponent(app Application) bool {
	_, ok := app.(HTTPComponent)
	return ok
}
func ImplementsWebsocketComponent(app Application) bool {
	_, ok := app.(WebsocketComponent)
	return ok
}

func ImplementsCommandHandlerComponent(app Application) bool {
	_, ok := app.(CommandHandlerComponent)
	return ok
}

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type CommandHandlerComponent interface {
	GetCommandHandlers() map[string]CommandHandler
}

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type SystemgeComponent interface {
	GetAsyncMessageHandlers() map[string]AsyncMessageHandler
	GetSyncMessageHandlers() map[string]SyncMessageHandler
	OnStart(*Node) error
	OnStop(*Node) error
	GetSystemgeConfig() Config.Systemge
}
type AsyncMessageHandler func(*Node, *Message.Message) error
type SyncMessageHandler func(*Node, *Message.Message) (string, error)
type CommandHandler func(*Node, []string) error

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type HTTPComponent interface {
	GetHTTPRequestHandlers() map[string]Http.RequestHandler
	GetHTTPComponentConfig() Config.HTTP
}

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type WebsocketComponent interface {
	GetWebsocketMessageHandlers() map[string]WebsocketMessageHandler
	OnConnectHandler(*Node, *WebsocketClient)
	OnDisconnectHandler(*Node, *WebsocketClient)
	GetWebsocketComponentConfig() Config.Websocket
}
type WebsocketMessageHandler func(*Node, *WebsocketClient, *Message.Message) error
