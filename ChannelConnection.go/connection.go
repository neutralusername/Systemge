package ChannelConnection

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type ConnectionChannel[T any] chan *ConnectionRequest[T]

type ConnectionRequest[T any] struct {
	messageChannel *MessageChannel[T]
	response       chan bool
}

type MessageChannel[T any] struct {
	sendChannel           chan T
	receiveChannelChannel chan T
}

// implements SystemgeConnection
type ChannelConnection[T any] struct {
	instanceId string

	closed       bool
	closedMutex  sync.Mutex
	closeChannel chan bool

	messageChannel MessageChannel[T]

	readRoutine *Tools.Routine

	writeMutex sync.Mutex
	readMutex  sync.RWMutex

	// metrics

	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
}

func New[T any](messageChannel MessageChannel[T]) *ChannelConnection[T] {

	client := &ChannelConnection[T]{
		closeChannel:   make(chan bool),
		instanceId:     Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		messageChannel: messageChannel,
	}

	return client
}

func (client *ChannelConnection[T]) GetStatus() int {
	client.closedMutex.Lock()
	defer client.closedMutex.Unlock()
	if client.closed {
		return Status.Stopped
	} else {
		return Status.Started
	}
}

func (client *ChannelConnection[T]) GetInstanceId() string {
	return client.instanceId
}

// GetCloseChannel returns a channel that will be closed when the connection is closed.
// Blocks until the connection is closed.
// This can be used to trigger an event when the connection is closed.
func (connclientction *ChannelConnection[T]) GetCloseChannel() <-chan bool {
	return connclientction.closeChannel
}

func (client *ChannelConnection[T]) GetAddress() string {
	return client.instanceId
}
