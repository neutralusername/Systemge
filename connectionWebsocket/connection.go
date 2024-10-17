package connectionWebsocket

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/constants"
	"github.com/neutralusername/Systemge/status"
	"github.com/neutralusername/Systemge/tools"
)

// implements SystemgeConnection
type WebsocketConnection struct {
	instanceId string

	websocketConn *websocket.Conn

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan struct{}

	writeMutex sync.Mutex
	readMutex  sync.RWMutex

	// metrics

	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New(websocketConn *websocket.Conn) (*WebsocketConnection, error) {
	if websocketConn == nil {
		return nil, errors.New("websocketConn is nil")
	}

	connection := &WebsocketConnection{
		websocketConn: websocketConn,
		closeChannel:  make(chan struct{}),
		instanceId:    tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),
	}

	return connection, nil
}

func (connection *WebsocketConnection) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return status.Stopped
	} else {
		return status.Started
	}
}

func (connection *WebsocketConnection) GetInstanceId() string {
	return connection.instanceId
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *WebsocketConnection) GetCloseChannel() <-chan struct{} {
	return connection.closeChannel
}

func (connection *WebsocketConnection) GetAddress() string {
	return connection.websocketConn.RemoteAddr().String()
}
