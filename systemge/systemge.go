package systemge

import (
	"github.com/neutralusername/systemge/tools"
)

type AcceptHandler[T any] func(Connection[T])
type AcceptHandlerWithError[T any] func(Connection[T]) error

type ReadHandler[T any] func(T, Connection[T])
type ReadHandlerWithResult[T any] func(T, Connection[T]) (T, error)
type ReadHandlerWithError[T any] func(T, Connection[T]) error

type AsyncMessageHandler[T any, P any] func(Connection[T], P)
type AsyncMessageHandlers[T any, P any] map[string]AsyncMessageHandler[T, P]

type SyncMessageHandler[T any, P any] func(Connection[T], P) (T, error)
type SyncMessageHandlers[T any, P any] map[string]SyncMessageHandler[T, P]

type Connector[T any] interface {
	Connect(int64) (Connection[T], error)
}

type Listener[T any] interface {
	Start() error
	Stop() error

	GetConnector() Connector[T]

	GetInstanceId() string
	GetSessionId() string
	GetName() string
	GetStatus() int
	GetStopChannel() <-chan struct{}

	Accept(int64) (Connection[T], error)
	SetAcceptDeadline(int64)

	GetDefaultCommands() tools.CommandHandlers

	CheckMetrics() tools.MetricsTypes
	GetMetrics() tools.MetricsTypes
}

type Connection[T any] interface {
	Close() error

	GetInstanceId() string
	GetAddress() string
	GetStatus() int
	GetCloseChannel() <-chan struct{}

	Read(int64) (T, error)
	SetReadDeadline(int64)

	Write(T, int64) error
	SetWriteDeadline(int64)

	GetDefaultCommands() tools.CommandHandlers

	GetMetrics() tools.MetricsTypes
	CheckMetrics() tools.MetricsTypes
}
