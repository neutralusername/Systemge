package systemge

import "github.com/neutralusername/systemge/tools"

type AsyncMessageHandler[T any] func(Connection[T], tools.IMessage)
type AsyncMessageHandlers[T any] map[string]AsyncMessageHandler[T]

type SyncMessageHandler[T any] func(Connection[T], tools.IMessage) (string, error)
type SyncMessageHandlers[T any] map[string]SyncMessageHandler[T]

func NewAsyncMessageHandlers[T any]() AsyncMessageHandlers[T] {
	return make(AsyncMessageHandlers[T])
}

func (map1 AsyncMessageHandlers[T]) Merge(map2 AsyncMessageHandlers[T]) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 AsyncMessageHandlers[T]) Add(key string, value AsyncMessageHandler[T]) {
	map1[key] = value
}

func (map1 AsyncMessageHandlers[T]) Remove(key string) {
	delete(map1, key)
}

func (map1 AsyncMessageHandlers[T]) Get(key string) (AsyncMessageHandler[T], bool) {
	value, ok := map1[key]
	return value, ok
}

func NewSyncMessageHandlers[T any]() SyncMessageHandlers[T] {
	return make(SyncMessageHandlers[T])
}

func (map1 SyncMessageHandlers[T]) Merge(map2 SyncMessageHandlers[T]) {
	for key, value := range map2 {
		map1[key] = value
	}
}

func (map1 SyncMessageHandlers[T]) Add(key string, value SyncMessageHandler[T]) {
	map1[key] = value
}

func (map1 SyncMessageHandlers[T]) Remove(key string) {
	delete(map1, key)
}

func (map1 SyncMessageHandlers[T]) Get(key string) (SyncMessageHandler[T], bool) {
	value, ok := map1[key]
	return value, ok
}
