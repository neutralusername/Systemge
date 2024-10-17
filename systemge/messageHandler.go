package systemge

import "github.com/neutralusername/systemge/tools"

type AsyncMessageHandler[B any] func(Connection[B], tools.IMessage)
type AsyncMessageHandlers[B any] map[string]AsyncMessageHandler[B]

type SyncMessageHandler[B any] func(Connection[B], tools.IMessage) (string, error)
type SyncMessageHandlers[B any] map[string]SyncMessageHandler[B]

func NewAsyncMessageHandlers[B any]() AsyncMessageHandlers[B] {
	return make(AsyncMessageHandlers[B])
}

func (map1 AsyncMessageHandlers[B]) Merge(map2 AsyncMessageHandlers[B]) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 AsyncMessageHandlers[B]) Add(key string, value AsyncMessageHandler[B]) {
	map1[key] = value
}

func (map1 AsyncMessageHandlers[B]) Remove(key string) {
	delete(map1, key)
}

func (map1 AsyncMessageHandlers[B]) Get(key string) (AsyncMessageHandler[B], bool) {
	value, ok := map1[key]
	return value, ok
}

func NewSyncMessageHandlers[B any]() SyncMessageHandlers[B] {
	return make(SyncMessageHandlers[B])
}

func (map1 SyncMessageHandlers[B]) Merge(map2 SyncMessageHandlers[B]) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 SyncMessageHandlers[B]) Add(key string, value SyncMessageHandler[B]) {
	map1[key] = value
}

func (map1 SyncMessageHandlers[B]) Remove(key string) {
	delete(map1, key)
}

func (map1 SyncMessageHandlers[B]) Get(key string) (SyncMessageHandler[B], bool) {
	value, ok := map1[key]
	return value, ok
}
