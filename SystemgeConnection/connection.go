package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Metrics"
)

type SystemgeConnection[B any] interface {
	Close() error

	GetInstanceId() string
	GetAddress() string
	GetStatus() int
	GetCloseChannel() <-chan struct{}

	/*
		StartReadRoutine(uint32, int64, int64, Tools.ReadHandler[O, C]) error
		StopReadRoutine() error
		IsReadRoutineRunning() bool
	*/

	// SetReadLimit(int64)

	Read(int64) (B, error)
	SetReadDeadline(int64)

	Write(B, int64) error
	SetWriteDeadline(int64)

	GetDefaultCommands() Commands.Handlers

	GetMetrics() Metrics.MetricsTypes
	CheckMetrics() Metrics.MetricsTypes
}
