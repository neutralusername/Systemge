package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Metrics"
)

type SystemgeConnection interface {
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

	Read() ([]byte, error)
	ReadTimeout(uint64) ([]byte, error)

	Write([]byte) error
	WriteTimeout([]byte, uint64) error

	GetDefaultCommands() Commands.Handlers

	GetMetrics() Metrics.MetricsTypes
	CheckMetrics() Metrics.MetricsTypes
}
