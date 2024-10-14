package TcpConnection

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type TcpConnection struct {
	config     *Config.TcpSystemgeConnection
	netConn    net.Conn
	randomizer *Tools.Randomizer

	readMutex  sync.RWMutex
	writeMutex sync.Mutex

	readRoutine *Tools.Routine

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan bool
	waitGroup    sync.WaitGroup

	messageReceiver *Tcp.BufferedMessageReader

	// metrics
	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New(config *Config.TcpSystemgeConnection, netConn net.Conn) (*TcpConnection, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if netConn == nil {
		return nil, errors.New("netConn is nil")
	}

	connection := &TcpConnection{
		config:          config,
		netConn:         netConn,
		messageReceiver: Tcp.NewBufferedMessageReader(netConn, config.IncomingMessageByteLimit, config.TcpReceiveTimeoutMs, config.TcpBufferBytes),
		randomizer:      Tools.NewRandomizer(config.RandomizerSeed),
		closeChannel:    make(chan bool),
	}
	if config.TcpBufferBytes <= 0 {
		config.TcpBufferBytes = 1024 * 4
	}

	connection.waitGroup.Add(1)
	return connection, nil
}

func (connection *TcpConnection) Close() error {
	if !connection.closedMutex.TryLock() {
		return errors.New("connection already closing")
	}
	defer connection.closedMutex.Unlock()

	if connection.closed {
		return errors.New("connection already closed")
	}

	connection.closed = true
	connection.netConn.Close()
	close(connection.closeChannel)
	connection.waitGroup.Wait()

	return nil
}

func (connection *TcpConnection) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return Status.Stopped
	} else {
		return Status.Started
	}
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *TcpConnection) GetCloseChannel() <-chan bool {
	return connection.closeChannel
}

func (connection *TcpConnection) GetAddress() string {
	return connection.netConn.RemoteAddr().String()
}
