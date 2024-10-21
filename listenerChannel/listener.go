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

type ChannelListener[D any] struct {
	name string

	instanceId string
	sessionId  string

	statusMutex sync.Mutex
	status      int
	stopChannel chan struct{}

	connectionChannel chan *connectionChannel.ConnectionRequest[D]
	timeout           *tools.Timeout
	mutex             sync.Mutex

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
}

func New[D any](name string) (systemge.Listener[D, systemge.Connection[D]], error) {
	listener := &ChannelListener[D]{
		name:       name,
		status:     status.Stopped,
		instanceId: tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),
	}

	return listener, nil
}

func (listener *ChannelListener[D]) GetConnectionChannel() chan<- *connectionChannel.ConnectionRequest[D] {
	return listener.connectionChannel
}

func (listener *ChannelListener[D]) GetStopChannel() <-chan struct{} {
	return listener.stopChannel
}

func (listener *ChannelListener[D]) GetInstanceId() string {
	return listener.instanceId
}

func (listener *ChannelListener[D]) GetSessionId() string {
	return listener.sessionId
}

func (listener *ChannelListener[D]) GetStatus() int {
	return listener.status
}

func (server *ChannelListener[D]) GetName() string {
	return server.name
}
