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

	SetReadDeadline(uint64)
	SetWriteDeadline(uint64)

	Read() (B, error)
	ReadTimeout(uint64) (B, error)

	Write(B) error
	WriteTimeout(B, uint64) error

	GetDefaultCommands() Commands.Handlers

	GetMetrics() Metrics.MetricsTypes
	CheckMetrics() Metrics.MetricsTypes
}
