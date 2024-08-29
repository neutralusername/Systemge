package SystemgeConnection

import "github.com/neutralusername/Systemge/Message"

type AsyncMessageHandler func(*SystemgeConnection, *Message.Message)
type AsyncMessageHandlers map[string]AsyncMessageHandler

type SyncMessageHandler func(*SystemgeConnection, *Message.Message) (string, error)
type SyncMessageHandlers map[string]SyncMessageHandler

func NewAsyncMessageHandlers() AsyncMessageHandlers {
	return make(AsyncMessageHandlers)
}

func (map1 AsyncMessageHandlers) Merge(map2 AsyncMessageHandlers) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 AsyncMessageHandlers) Add(key string, value AsyncMessageHandler) {
	map1[key] = value
}

func (map1 AsyncMessageHandlers) Remove(key string) {
	delete(map1, key)
}

func (map1 AsyncMessageHandlers) Get(key string) (AsyncMessageHandler, bool) {
	value, ok := map1[key]
	return value, ok
}

func NewSyncMessageHandlers() SyncMessageHandlers {
	return make(SyncMessageHandlers)
}

func (map1 SyncMessageHandlers) Merge(map2 SyncMessageHandlers) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 SyncMessageHandlers) Add(key string, value SyncMessageHandler) {
	map1[key] = value
}

func (map1 SyncMessageHandlers) Remove(key string) {
	delete(map1, key)
}

func (map1 SyncMessageHandlers) Get(key string) (SyncMessageHandler, bool) {
	value, ok := map1[key]
	return value, ok
}

type MessageHandler interface {
	HandleAsyncMessage(connection *SystemgeConnection, message *Message.Message) error
	HandleSyncRequest(connection *SystemgeConnection, message *Message.Message) (string, error)
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

	GetAsyncMessagesHandled() uint64
	RetrieveAsyncMessagesHandled() uint64

	GetSyncRequestsHandled() uint64
	RetrieveSyncRequestsHandled() uint64

	GetUnknownTopicsReceived() uint64
	RetrieveUnknownTopicsReceived() uint64
}
