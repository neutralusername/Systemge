package Application

import (
	"Systemge/Message"
)

type Application interface {
	GetAsyncMessageHandlers() map[string]AsyncMessageHandler
	GetSyncMessageHandlers() map[string]SyncMessageHandler
	GetCustomCommandHandlers() map[string]CustomCommandHandler
	OnStart() error
	OnStop() error
}
type AsyncMessageHandler func(*Message.Message) error
type SyncMessageHandler func(*Message.Message) (string, error)
type CustomCommandHandler func([]string) error
