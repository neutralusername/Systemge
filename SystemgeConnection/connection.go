package SystemgeConnection

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
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

	syncResponseChannels map[string]chan *Message.Message
	syncMutex            sync.Mutex

	stopChannel chan bool

	// metrics
	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64

	asyncMessagesSent atomic.Uint32
	syncRequestsSent  atomic.Uint32

	syncSuccessResponsesReceived atomic.Uint32
	syncFailureResponsesReceived atomic.Uint32
	noSyncResponseReceived       atomic.Uint32
}

func New(config *Config.SystemgeConnection, netConn net.Conn, name string) *SystemgeConnection {
	connection := &SystemgeConnection{
		name:        name,
		config:      config,
		netConn:     netConn,
		randomizer:  Tools.NewRandomizer(config.RandomizerSeed),
		stopChannel: make(chan bool),
	}
	return connection
}

func (connection *SystemgeConnection) Close() {
	connection.netConn.Close()
	close(connection.stopChannel)
}

func (connection *SystemgeConnection) GetName() string {
	return connection.name
}
