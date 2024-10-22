package systemge

import (
	"github.com/neutralusername/systemge/tools"
)

type Server[D any] interface {
	GetReadHandler() tools.ReadHandler[D, Connection[D]]
	GetAcceptHandler() tools.AcceptHandlerWithError[Connection[D]]
	SetReadHandler(tools.ReadHandler[D, Connection[D]])
	SetAcceptHandler(tools.AcceptHandlerWithError[Connection[D]])
}

type Connector[D any, C Connection[D]] interface {
	Connect(int64) (C, error)
}

type Listener[D any, C Connection[D]] interface {
	Start() error
	Stop() error

	GetConnector() Connector[D, C]

	GetInstanceId() string
	GetSessionId() string
	GetName() string
	GetStatus() int
	GetStopChannel() <-chan struct{}

	Accept(int64) (C, error)
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
