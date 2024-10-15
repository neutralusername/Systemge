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
	GetCloseChannel() <-chan bool

	/* 	StartReadRoutine(uint32, int64, int64, Tools.ReadHandler[O, C]) error
	   	StopReadRoutine() error
	   	IsReadRoutineRunning() bool */

	SetReadDeadline(int64)
	SetWriteDeadline(int64)

	Read(int64) (B, error)

	Write(B, int64) error

	GetDefaultCommands() Commands.Handlers

	GetMetrics() Metrics.MetricsTypes
	CheckMetrics() Metrics.MetricsTypes
}
