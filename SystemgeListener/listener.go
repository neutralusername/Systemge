package SystemgeListener

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeListener interface {
	AcceptConnection(serverName string, connectionConfig *Config.TcpSystemgeConnection) (SystemgeConnection.SystemgeConnection, error)
	Close() error
	GetAcceptedConnections() uint64
	GetBlacklist() *Tools.AccessControlList
	GetDefaultCommands() Commands.Handlers
	GetConnectionAttempts() uint64
	GetFailedConnections() uint64
	GetRejectedConnections() uint64
	GetWhitelist() *Tools.AccessControlList
	RetrieveAcceptedConnections() uint64
	RetrieveConnectionAttempts() uint64
	RetrieveFailedConnections() uint64
	RetrieveRejectedConnections() uint64
}
