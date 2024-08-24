package SystemgeMessageHandler

import "github.com/neutralusername/Systemge/Message"

type AsyncMessageHandler func(*Message.Message)
type AsyncMessageHandlers map[string]AsyncMessageHandler

type SyncMessageHandler func(*Message.Message) (string, error)
type SyncMessageHandlers map[string]SyncMessageHandler

type MessageHandler interface {
	HandleAsyncMessage(message *Message.Message) error
	HandleSyncRequest(message *Message.Message) (string, error)
	AddAsyncMessageHandler(topic string, handler func(*Message.Message))
	AddSyncMessageHandler(topic string, handler func(*Message.Message) (string, error))
	RemoveAsyncMessageHandler(topic string)
	RemoveSyncMessageHandler(topic string)
	SetUnknownAsyncHandler(handler func(*Message.Message))
	SetUnknownSyncHandler(handler func(*Message.Message) (string, error))
	GetAsyncMessageHandler(topic string) func(*Message.Message)
	GetSyncMessageHandler(topic string) func(*Message.Message) (string, error)

	GetAsyncMessagesHandled() uint64
	RetrieveAsyncMessagesHandled() uint64

	GetSyncRequestsHandled() uint64
	RetrieveSyncRequestsHandled() uint64

	GetUnknownTopicsReceived() uint64
	RetrieveUnknownTopicsReceived() uint64
}
