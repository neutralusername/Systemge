package Client

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/Config"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/SessionManager"
	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

type Client struct {
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

func New(name string, config *Config.SystemgeClient, eventHandler Event.Handler) (*Client, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpClientConfigs == nil {
		return nil, errors.New("config.TcpClientConfigs is nil")
	}
	if config.TcpSystemgeConnectionConfig == nil {
		return nil, errors.New("config.TcpSystemgeConnectionConfig is nil")
	}

	client := &Client{
		name:   name,
		config: config,

		sessionManager: SessionManager.New(name+"_sessionManager", config.SessionManagerConfig, eventHandler),

		instanceId: tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),

		eventHandler: eventHandler,
	}

	return client, nil
}

func (client *Client) GetName() string {
	return client.name
}

func (client *Client) GetStatus() int {
	return client.status
}

func (server *Client) GetInstanceId() string {
	return server.instanceId
}

func (server *Client) GetSessionId() string {
	return server.sessionId
}

func (server *Client) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetContext())
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *Client) GetContext() Event.Context {
	return Event.Context{
		Event.ServiceType:       Event.SystemgeClient,
		Event.ServiceName:       server.name,
		Event.ServiceStatus:     status.ToString(server.status),
		Event.Function:          Event.GetCallerFuncName(2),
		Event.ServiceInstanceId: server.instanceId,
		Event.SessionId:         server.sessionId,
	}
}
