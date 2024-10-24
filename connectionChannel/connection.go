package connectionChannel

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

type ConnectionRequest[T any] struct {
	SendToListener      chan T
	ReceiveFromListener chan T
}

// implements SystemgeConnection
type ChannelConnection[T any] struct {
	instanceId string

	receiveChannel chan T
	sendChannel    chan T

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan struct{}

	writeMutex   sync.Mutex
	writeTimeout *tools.Timeout

	readMutex   sync.RWMutex
	readTimeout *tools.Timeout

	// metrics

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New[T any](receiveChannel chan T, sendChannel chan T) *ChannelConnection[T] {
	connection := &ChannelConnection[T]{
		closeChannel:   make(chan struct{}),
		instanceId:     tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),
		receiveChannel: receiveChannel,
		sendChannel:    sendChannel,
	}

	return connection
}

func (connection *ChannelConnection[T]) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return status.Stopped
	} else {
		return status.Started
	}
}

func (connection *ChannelConnection[T]) GetInstanceId() string {
	return connection.instanceId
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *ChannelConnection[T]) GetCloseChannel() <-chan struct{} {
	return connection.closeChannel
}

func (connection *ChannelConnection[T]) GetAddress() string {
	return connection.instanceId
}
