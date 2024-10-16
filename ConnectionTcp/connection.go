package ConnectionTcp

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

// implements SystemgeConnection
type TcpConnection struct {
	config     *Config.TcpSystemgeConnection
	instanceId string

	netConn         net.Conn
	messageReceiver *BufferedMessageReader

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan struct{}

	readRoutine *Tools.Routine

	readMutex  sync.RWMutex
	writeMutex sync.Mutex

	// metrics
	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New(netConn net.Conn) (*TcpConnection, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if netConn == nil {
		return nil, errors.New("netConn is nil")
	}
	if config.TcpBufferBytes <= 0 {
		config.TcpBufferBytes = 1024 * 4
	}

	connection := &TcpConnection{
		config:          config,
		netConn:         netConn,
		messageReceiver: NewBufferedMessageReader(netConn, config.IncomingMessageByteLimit, config.TcpReceiveTimeoutNs, config.TcpBufferBytes),
		closeChannel:    make(chan struct{}),
		instanceId:      Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
	}

	return connection, nil
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
func (connection *TcpConnection) GetCloseChannel() <-chan struct{} {
	return connection.closeChannel
}

func (connection *TcpConnection) GetAddress() string {
	return connection.netConn.RemoteAddr().String()
}

func (connection *TcpConnection) GetInstanceId() string {
	return connection.instanceId
}
