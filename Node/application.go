package Node

import (
	"net/http"

	"github.com/neutralusername/Systemge/Message"
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
	_, ok := app.(CommandComponent)
	return ok
}

func ImplementsOnStartComponent(app Application) bool {
	_, ok := app.(OnStartComponent)
	return ok
}

func ImplementsOnStopComponent(app Application) bool {
	_, ok := app.(OnStopComponent)
	return ok
}

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type CommandComponent interface {
	GetCommandHandlers() map[string]CommandHandler
}
type CommandHandler func(*Node, []string) (string, error)

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type OnStartComponent interface {
	OnStart(*Node) error
}

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type OnStopComponent interface {
	OnStop(*Node) error
}

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type SystemgeComponent interface {
	GetAsyncMessageHandlers() map[string]AsyncMessageHandler
	GetSyncMessageHandlers() map[string]SyncMessageHandler
}
type AsyncMessageHandler func(*Node, *Message.Message) error
type SyncMessageHandler func(*Node, *Message.Message) (string, error)

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type HTTPComponent interface {
	GetHTTPMessageHandlers() map[string]http.HandlerFunc
}

// if a struct embeds this interface and does not implement its methods, it will cause a runtime panic if passed to a node
type WebsocketComponent interface {
	GetWebsocketMessageHandlers() map[string]WebsocketMessageHandler
	OnConnectHandler(*Node, *WebsocketClient)
	OnDisconnectHandler(*Node, *WebsocketClient)
}
type WebsocketMessageHandler func(*Node, *WebsocketClient, *Message.Message) error
