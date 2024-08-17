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

	receiver *SystemgeReceiver

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex

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

func New(config *Config.SystemgeConnection, netConn net.Conn, name string, messageHandler *SystemgeMessageHandler) *SystemgeConnection {
	connection := &SystemgeConnection{
		name:         name,
		config:       config,
		netConn:      netConn,
		randomizer:   Tools.NewRandomizer(config.RandomizerSeed),
		closeChannel: make(chan bool),
	}
	if config.ReceiverConfig != nil {
		receiver := NewReceiver(config.ReceiverConfig, connection, messageHandler)
		connection.receiver = receiver
	}
	return connection
}

func (connection *SystemgeConnection) Close() {
	connection.netConn.Close()
	if connection.receiver != nil {
		connection.receiver.Close()
	}
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
