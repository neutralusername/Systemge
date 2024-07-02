package Node

import (
	"Systemge/Config"
	"Systemge/Message"
	"net/http"
)

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type Application interface {
	GetAsyncMessageHandlers() map[string]AsyncMessageHandler
	GetSyncMessageHandlers() map[string]SyncMessageHandler
	GetCustomCommandHandlers() map[string]CustomCommandHandler
	OnStart(*Node) error
	OnStop(*Node) error
	GetApplicationConfig() Config.Application
}
type AsyncMessageHandler func(*Node, *Message.Message) error
type SyncMessageHandler func(*Node, *Message.Message) (string, error)
type CustomCommandHandler func(*Node, []string) error

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type HTTPComponent interface {
	GetHTTPRequestHandlers() map[string]HTTPRequestHandler
	GetHTTPComponentConfig() Config.HTTP
}
type HTTPRequestHandler func(w http.ResponseWriter, r *http.Request)

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type WebsocketComponent interface {
	GetWebsocketMessageHandlers() map[string]WebsocketMessageHandler
	OnConnectHandler(*Node, *WebsocketClient)
	OnDisconnectHandler(*Node, *WebsocketClient)
	GetWebsocketComponentConfig() Config.Websocket
}
type WebsocketMessageHandler func(*Node, *WebsocketClient, *Message.Message) error

func ImplementsApplication(obj interface{}) bool {
	_, ok := obj.(Application)
	return ok
}

func ImplementsHTTPComponent(obj interface{}) bool {
	_, ok := obj.(HTTPComponent)
	return ok
}

func ImplementsWebsocketComponent(obj interface{}) bool {
	_, ok := obj.(WebsocketComponent)
	return ok
}
