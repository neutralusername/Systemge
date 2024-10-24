package systemge

type AsyncMessageHandler[T any, P any] func(Connection[T], P)
type AsyncMessageHandlers[T any, P any] map[string]AsyncMessageHandler[T, P]

type SyncMessageHandler[T any, P any] func(Connection[T], P) (T, error)
type SyncMessageHandlers[T any, P any] map[string]SyncMessageHandler[T, P]
