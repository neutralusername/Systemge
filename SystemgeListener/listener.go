package SystemgeListener

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/TcpConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeListener interface {
	AcceptConnection(serverName string, connectionConfig *Config.SystemgeConnection) (*TcpConnection.TcpConnection, error)
	Close()
	GetAcceptedConnections() uint64
	GetBlacklist() *Tools.AccessControlList
	GetConnectionAttempts() uint64
	GetFailedConnections() uint64
	GetRejectedConnections() uint64
	GetWhitelist() *Tools.AccessControlList
	RetrieveAcceptedConnections() uint64
	RetrieveConnectionAttempts() uint64
	RetrieveFailedConnections() uint64
	RetrieveRejectedConnections() uint64
}
