package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Message"
)

type MessageHandler interface {
	HandleAsyncMessage(connection SystemgeConnection, message *Message.Message) error
	HandleSyncRequest(connection SystemgeConnection, message *Message.Message) (string, error)
	AddAsyncMessageHandler(topic string, handler AsyncMessageHandler)
	AddSyncMessageHandler(topic string, handler SyncMessageHandler)
	RemoveAsyncMessageHandler(topic string)
	RemoveSyncMessageHandler(topic string)
	SetUnknownAsyncHandler(handler AsyncMessageHandler)
	SetUnknownSyncHandler(handler SyncMessageHandler)
	GetAsyncMessageHandler(topic string) AsyncMessageHandler
	GetSyncMessageHandler(topic string) SyncMessageHandler
	GetAsyncTopics() []string
	GetSyncTopics() []string

	CheckMetrics() map[string]uint64
	GetMetrics() map[string]uint64

	CheckAsyncMessagesHandled() uint64
	GetAsyncMessagesHandled() uint64

	CheckSyncRequestsHandled() uint64
	GetSyncRequestsHandled() uint64

	CheckUnknownTopicsReceived() uint64
	GetUnknownTopicsReceived() uint64
}
