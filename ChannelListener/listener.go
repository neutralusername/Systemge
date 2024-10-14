package WebsocketListener

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type ChannelListener[T any] struct {
	config *Config.ChannelListener
	name   string

	instanceId string
	sessionId  string

	statusMutex sync.Mutex
	status      int

	acceptMutex sync.RWMutex

	stopChannel chan struct{}

	acceptRoutine *Tools.Routine

	connectionChannel chan (connectionRequest[T])

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
	ClientsRejected atomic.Uint64
}

type connectionRequest[T any] struct {
	errChannel        chan error
	toClientChannel   chan T
	fromClientChannel chan T
}

func New[T any](name string, config *Config.ChannelListener) (*ChannelListener[T], error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	listener := &ChannelListener[T]{
		name:              name,
		config:            config,
		status:            Status.Stopped,
		instanceId:        Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
		connectionChannel: make(chan (connectionRequest[T])),
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
