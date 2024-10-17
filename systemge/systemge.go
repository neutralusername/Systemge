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

	Accept(timeoutNs int64) (C, error)

	GetDefaultCommands() tools.Handlers

	CheckMetrics() tools.MetricsTypes
	GetMetrics() tools.MetricsTypes
}

type Connection[B any] interface {
	Close() error

	GetInstanceId() string
	GetAddress() string
	GetStatus() int
	GetCloseChannel() <-chan struct{}

	// SetReadLimit(int64) would be nice but redundant for channel communication. integrate this config somehow into tcp and websocket connections

	Read(int64) (B, error)
	SetReadDeadline(int64)

	Write(B, int64) error
	SetWriteDeadline(int64)

	GetDefaultCommands() tools.Handlers

	GetMetrics() tools.MetricsTypes
	CheckMetrics() tools.MetricsTypes
}
