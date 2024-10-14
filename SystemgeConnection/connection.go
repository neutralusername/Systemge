package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Metrics"
)

type SystemgeConnection interface {
	GetAddress() string
	GetStatus() int
	Close() error
	GetCloseChannel() <-chan bool

	Write([]byte) error

	GetDefaultCommands() Commands.Handlers

	GetMetrics() Metrics.MetricsTypes
	CheckMetrics() Metrics.MetricsTypes
}
