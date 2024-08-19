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

	closed      bool
	closedMutex sync.Mutex

	tcpBuffer []byte

	syncResponseChannels map[string]chan *Message.Message
	syncMutex            sync.Mutex

	closeChannel chan bool

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
		name:                 name,
		config:               config,
		netConn:              netConn,
		randomizer:           Tools.NewRandomizer(config.RandomizerSeed),
		closeChannel:         make(chan bool),
		syncResponseChannels: make(map[string]chan *Message.Message),
	}
	return connection
}

func (connection *SystemgeConnection) Close() {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return
	}
	connection.closed = true
	connection.netConn.Close()
	close(connection.closeChannel)
}

func (connection *SystemgeConnection) GetName() string {
	return connection.name
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *SystemgeConnection) GetCloseChannel() <-chan bool {
	return connection.closeChannel
}

func (connection *SystemgeConnection) GetAddress() string {
	return connection.netConn.RemoteAddr().String()
}
