package SystemgeListener

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type SystemgeListener interface {
	AcceptConnection(serverName string, connectionConfig *Config.TcpSystemgeConnection) (SystemgeConnection.SystemgeConnection, error)
	Close() error
	GetDefaultCommands() Commands.Handlers
	CheckAcceptedConnections() uint64
	CheckConnectionAttempts() uint64
	CheckFailedConnections() uint64
	CheckRejectedConnections() uint64
	GetAcceptedConnections() uint64
	GetConnectionAttempts() uint64
	GetFailedConnections() uint64
	GetRejectedConnections() uint64
	GetMetrics() map[string]*Metrics.Metrics
	CheckMetrics() map[string]*Metrics.Metrics
}
