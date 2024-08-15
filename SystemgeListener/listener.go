package SystemgeListener

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeListener struct {
	status      int
	statusMutex sync.Mutex

	config *Config.SystemgeListener

	tcpServer *Tcp.Server

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer

	clients map[string]*SystemgeConnection.SystemgeConnection

	stopChannel                            chan bool //closing of this channel initiates the stop of the systemge component
	clientConnectionStopChannelStopChannel chan bool //closing of this channel indicates that the client connection handler has stopped
	allClientConnectionsStoppedChannel     chan bool //closing of this channel indicates that all client connections have stopped

	mutex sync.Mutex

	// metrics
	connectionAttempts  atomic.Uint32
	rejectedConnections atomic.Uint32
	acceptedConnections atomic.Uint32
}
