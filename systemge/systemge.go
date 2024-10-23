package systemge

import (
	"github.com/neutralusername/systemge/tools"
)

type AcceptHandler[D any] func(Connection[D])
type AcceptHandlerWithError[D any] func(Connection[D]) error

type ReadHandler[D any] func(D, Connection[D])
type ReadHandlerWithResult[D any] func(D, Connection[D]) (D, error)
type ReadHandlerWithError[D any] func(D, Connection[D]) error

type Connector[D any] interface {
	Connect(int64) (Connection[D], error)
}

type Listener[D any] interface {
	Start() error
	Stop() error

	GetConnector() Connector[D]

	GetInstanceId() string
	GetSessionId() string
	GetName() string
	GetStatus() int
	GetStopChannel() <-chan struct{}

	Accept(int64) (Connection[D], error)
	SetAcceptDeadline(int64)

	GetDefaultCommands() tools.CommandHandlers

	CheckMetrics() tools.MetricsTypes
	GetMetrics() tools.MetricsTypes
}

type Connection[D any] interface {
	Close() error

	GetInstanceId() string
	GetAddress() string
	GetStatus() int
	GetCloseChannel() <-chan struct{}

	Read(int64) (D, error)
	SetReadDeadline(int64)

	Write(D, int64) error
	SetWriteDeadline(int64)

	GetDefaultCommands() tools.CommandHandlers

	GetMetrics() tools.MetricsTypes
	CheckMetrics() tools.MetricsTypes
}
