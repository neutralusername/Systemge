package Node

import (
	"github.com/neutralusername/Systemge/Message"
)

type Application interface{}

func ImplementsSystemgeServerComponent(app Application) bool {
	_, ok := app.(SystemgeServerComponent)
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
type SystemgeServerComponent interface {
	GetAsyncMessageHandlers() map[string]AsyncMessageHandler
	GetSyncMessageHandlers() map[string]SyncMessageHandler
}
type AsyncMessageHandler func(*Node, *Message.Message) error
type SyncMessageHandler func(*Node, *Message.Message) (string, error)
