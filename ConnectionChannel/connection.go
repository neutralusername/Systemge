package ConnectionChannel

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
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
	closeChannel chan bool

	readRoutine *Tools.Routine

	writeMutex sync.Mutex
	readMutex  sync.RWMutex

	// metrics

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New[T any](receiveChannel chan T, sendChannel chan T) *ChannelConnection[T] {

	connection := &ChannelConnection[T]{
		closeChannel:   make(chan bool),
		instanceId:     Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		receiveChannel: receiveChannel,
		sendChannel:    sendChannel,
	}

	return connection
}

func (connection *ChannelConnection[T]) GetStatus() int {
	connection.closedMutex.Lock()
	defer connection.closedMutex.Unlock()
	if connection.closed {
		return Status.Stopped
	} else {
		return Status.Started
	}
}

func (connection *ChannelConnection[T]) GetInstanceId() string {
	return connection.instanceId
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connection *ChannelConnection[T]) GetCloseChannel() <-chan bool {
	return connection.closeChannel
}

func (connection *ChannelConnection[T]) GetAddress() string {
	return connection.instanceId
}
