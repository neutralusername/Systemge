package systemge

import "github.com/neutralusername/systemge/tools"

type AsyncMessageHandler[D any] func(Connection[D], tools.IMessage)
type AsyncMessageHandlers[D any] map[string]AsyncMessageHandler[D]

type SyncMessageHandler[D any] func(Connection[D], tools.IMessage) (string, error)
type SyncMessageHandlers[D any] map[string]SyncMessageHandler[D]

func NewAsyncMessageHandlers[D any]() AsyncMessageHandlers[D] {
	return make(AsyncMessageHandlers[D])
}

func (map1 AsyncMessageHandlers[D]) Merge(map2 AsyncMessageHandlers[D]) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 AsyncMessageHandlers[D]) Add(key string, value AsyncMessageHandler[D]) {
	map1[key] = value
}

func (map1 AsyncMessageHandlers[D]) Remove(key string) {
	delete(map1, key)
}

func (map1 AsyncMessageHandlers[D]) Get(key string) (AsyncMessageHandler[D], bool) {
	value, ok := map1[key]
	return value, ok
}

func NewSyncMessageHandlers[D any]() SyncMessageHandlers[D] {
	return make(SyncMessageHandlers[D])
}

func (map1 SyncMessageHandlers[D]) Merge(map2 SyncMessageHandlers[D]) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 SyncMessageHandlers[D]) Add(key string, value SyncMessageHandler[D]) {
	map1[key] = value
}

func (map1 SyncMessageHandlers[D]) Remove(key string) {
	delete(map1, key)
}

func (map1 SyncMessageHandlers[D]) Get(key string) (SyncMessageHandler[D], bool) {
	value, ok := map1[key]
	return value, ok
}
