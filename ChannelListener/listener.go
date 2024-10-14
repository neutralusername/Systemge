package WebsocketListener

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/ChannelConnection.go"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type ChannelListener[T any] struct {
	name string

	instanceId string
	sessionId  string

	statusMutex sync.Mutex
	status      int

	acceptMutex sync.RWMutex

	stopChannel chan struct{}

	acceptRoutine *Tools.Routine

	connectionChannel ChannelConnection.ConnectionChannel[T]

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
	ClientsRejected atomic.Uint64
}

func New[T any](name string) (*ChannelListener[T], error) {
	listener := &ChannelListener[T]{
		name:              name,
		status:            Status.Stopped,
		instanceId:        Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		connectionChannel: make(ChannelConnection.ConnectionChannel[T]),
	}

	return listener, nil
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
