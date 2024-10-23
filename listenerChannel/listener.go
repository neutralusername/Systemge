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
	timeout           *tools.Timeout
	mutex             sync.Mutex

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
}

func New[T any](name string) (systemge.Listener[T], error) {
	listener := &ChannelListener[T]{
		name:       name,
		status:     status.Stopped,
		instanceId: tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),
	}

	return listener, nil
}

func (listener *ChannelListener[T]) GetConnector() systemge.Connector[T] {
	return &connector[T]{
		connChann: listener.connectionChannel,
	}
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
