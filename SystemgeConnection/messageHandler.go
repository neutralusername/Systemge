package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Message"
)

type AsyncMessageHandler func(SystemgeConnection, *Message.Message)
type AsyncMessageHandlers map[string]AsyncMessageHandler

type SyncMessageHandler func(SystemgeConnection, *Message.Message) (string, error)
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
