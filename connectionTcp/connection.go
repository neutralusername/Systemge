package connectionTcp

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

// implements SystemgeConnection
type TcpConnection struct {
	config     *configs.TcpBufferedReader
	instanceId string

	netConn           net.Conn
	tcpBufferedReader *tools.TcpBufferedReader

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan struct{}

	readMutex  sync.RWMutex
	writeMutex sync.Mutex

	lifeTimeout *tools.Timeout

	// metrics
	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New(config *configs.TcpBufferedReader, netConn net.Conn, lifetimeNs int64) (*TcpConnection, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if netConn == nil {
		return nil, errors.New("netConn is nil")
	}

	connection := &TcpConnection{
		config:            config,
		netConn:           netConn,
		tcpBufferedReader: tools.NewTcpBufferedReader(netConn, config),
		closeChannel:      make(chan struct{}),
		instanceId:        tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),
	}

	connection.lifeTimeout = tools.NewTimeout(
		lifetimeNs,
		func() {
			connection.Close()
		},
		false,
	)

	return connection, nil
}

func (connection *TcpConnection) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return status.Stopped
	} else {
		return status.Started
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

func (connection *TcpConnection) GetLifeTimeout() *tools.Timeout {
	return connection.lifeTimeout
}
