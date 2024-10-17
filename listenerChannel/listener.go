package listenerChannel

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type ChannelListener[T any] struct {
	name string

	instanceId string
	sessionId  string

	statusMutex sync.Mutex
	status      int
	stopChannel chan struct{}

	connectionChannel chan *connectionChannel.ConnectionRequest[T]

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
}

func New[B any](name string) (systemge.Listener[B, systemge.Connection[B]], error) {
	listener := &ChannelListener[B]{
		name:              name,
		status:            status.Stopped,
		instanceId:        tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),
		connectionChannel: make(chan *connectionChannel.ConnectionRequest[B]),
	}

	return listener, nil
}

func (listener *ChannelListener[T]) GetConnectionChannel() chan<- *connectionChannel.ConnectionRequest[T] {
	return listener.connectionChannel
}

func (listener *ChannelListener[T]) GetStopChannel() <-chan struct{} {
	return listener.stopChannel
}

func (listener *ChannelListener[T]) GetInstanceId() string {
	return listener.instanceId
}

func (listener *ChannelListener[T]) GetSessionId() string {
	return listener.sessionId
}

func (listener *ChannelListener[T]) GetStatus() int {
	return listener.status
}

func (server *ChannelListener[T]) GetName() string {
	return server.name
}
