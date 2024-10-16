package Systemge

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Metrics"
)

type Listener[B any] interface {
	Start() error
	Stop() error

	GetInstanceId() string
	GetSessionId() string
	GetName() string
	GetStatus() int
	GetStopChannel() <-chan struct{}

	Accept(timeoutNs int64) (Connection[B], error)

	GetDefaultCommands() Commands.Handlers

	CheckMetrics() Metrics.MetricsTypes
	GetMetrics() Metrics.MetricsTypes
}

type Connection[B any] interface {
	Close() error

	GetInstanceId() string
	GetAddress() string
	GetStatus() int
	GetCloseChannel() <-chan struct{}

	// SetReadLimit(int64) would be nice but redundant for channel communication

	Read(int64) (B, error)
	SetReadDeadline(int64)

	Write(B, int64) error
	SetWriteDeadline(int64)

	GetDefaultCommands() Commands.Handlers

	GetMetrics() Metrics.MetricsTypes
	CheckMetrics() Metrics.MetricsTypes
}
