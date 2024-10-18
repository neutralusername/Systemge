package systemge

import (
	"github.com/neutralusername/systemge/tools"
)

type Listener[B any, C Connection[B]] interface {
	Start() error
	Stop() error

	GetInstanceId() string
	GetSessionId() string
	GetName() string
	GetStatus() int
	GetStopChannel() <-chan struct{}

	AcceptChannel() <-chan C
	Accept(timeoutNs int64) (C, error)
	SetAcceptTimeout(int64)

	GetDefaultCommands() tools.CommandHandlers

	CheckMetrics() tools.MetricsTypes
	GetMetrics() tools.MetricsTypes
}

type Connection[B any] interface {
	Close() error

	GetInstanceId() string
	GetAddress() string
	GetStatus() int
	GetCloseChannel() <-chan struct{}

	ReadChannel() <-chan B
	Read(int64) (B, error)
	SetReadDeadline(int64)

	// WriteChannel() <-chan error
	Write(B, int64) error
	SetWriteDeadline(int64)

	GetDefaultCommands() tools.CommandHandlers

	GetMetrics() tools.MetricsTypes
	CheckMetrics() tools.MetricsTypes
}
