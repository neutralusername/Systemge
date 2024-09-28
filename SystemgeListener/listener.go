package SystemgeListener

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type SystemgeListener interface {
	AcceptConnection(connectionConfig *Config.TcpSystemgeConnection, eventHandler Event.Handler) (SystemgeConnection.SystemgeConnection, error)
	Close() error

	GetDefaultCommands() Commands.Handlers

	GetConnectionAttempts() uint64
	GetAcceptedConnectionAttempts() uint64
	GetFailedConnectionAttempts() uint64
	GetRejectedConnectionAttempts() uint64
	GetMetrics() Metrics.MetricsTypes

	CheckConnectionAttempts() uint64
	CheckAcceptedConnectionAttempts() uint64
	CheckFailedConnectionAttempts() uint64
	CheckRejectedConnectionAttempts() uint64
	CheckMetrics() Metrics.MetricsTypes
}
