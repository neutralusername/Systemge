package SystemgeClient

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/SessionManager"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeClient struct {
	name string

	instanceId string
	sessionId  string

	status      int
	statusMutex sync.RWMutex

	stopChannel chan bool
	waitGroup   sync.WaitGroup

	config *Config.SystemgeClient

	sessionManager *SessionManager.Manager

	eventHandler Event.Handler

	ongoingConnectionAttempts atomic.Int64

	// metrics

	connectionAttemptsFailed   atomic.Uint64
	connectionAttemptsRejected atomic.Uint64
	connectionAttemptsSuccess  atomic.Uint64
}

func New(name string, config *Config.SystemgeClient, eventHandler Event.Handler) (*SystemgeClient, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpClientConfigs == nil {
		return nil, errors.New("config.TcpClientConfigs is nil")
	}
	if config.TcpSystemgeConnectionConfig == nil {
		return nil, errors.New("config.TcpSystemgeConnectionConfig is nil")
	}

	client := &SystemgeClient{
		name:   name,
		config: config,

		sessionManager: SessionManager.New(name+"_sessionManager", config.SessionManagerConfig, eventHandler),

		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		eventHandler: eventHandler,
	}

	return client, nil
}

func (client *SystemgeClient) GetName() string {
	return client.name
}

func (client *SystemgeClient) GetStatus() int {
	return client.status
}

func (server *SystemgeClient) GetInstanceId() string {
	return server.instanceId
}

func (server *SystemgeClient) GetSessionId() string {
	return server.sessionId
}

func (server *SystemgeClient) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetContext())
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *SystemgeClient) GetContext() Event.Context {
	return Event.Context{
		Event.ServiceType:       Event.SystemgeClient,
		Event.ServiceName:       server.name,
		Event.ServiceStatus:     Status.ToString(server.status),
		Event.Function:          Event.GetCallerFuncName(2),
		Event.ServiceInstanceId: server.instanceId,
		Event.SessionId:         server.sessionId,
	}
}
