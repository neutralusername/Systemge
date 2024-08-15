package SystemgeConnection

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeConnection struct {
	name       string
	config     *Config.SystemgeConnection
	netConn    net.Conn
	randomizer *Tools.Randomizer

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex

	tcpBuffer []byte

	// metrics
	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64
}

func New(config *Config.SystemgeConnection, netConn net.Conn, name string) *SystemgeConnection {
	connection := &SystemgeConnection{
		name:       name,
		config:     config,
		netConn:    netConn,
		randomizer: Tools.NewRandomizer(config.RandomizerSeed),
	}
	return connection
}

func (connection *SystemgeConnection) GetName() string {
	return connection.name
}
