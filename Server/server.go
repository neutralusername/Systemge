package Server

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Systemge"
	"github.com/neutralusername/Systemge/Tools"
)

// ? == might not make it

type Server[B any, O any] struct {
	instanceId string
	sessionId  string
	name       string

	status      int
	statusMutex sync.RWMutex
	stopChannel chan bool
	waitGroup   sync.WaitGroup

	config *Config.Server

	acceptHandler    Tools.AcceptHandler[Systemge.Connection[B]]
	receptionHandler Tools.ReadHandler[O, Systemge.Connection[B]]

	listener Systemge.Listener[B]

	sessionManager         *Tools.SessionManager            // ?
	requestResponseManager *Tools.RequestResponseManager[B] // ?

	priorityTokenQueue *Tools.PriorityTokenQueue[O] // ?
	queueConsumer      *Tools.IQueueConsumer[O]     // ?

	topicManager  *Tools.TopicManager[O, O] // ?
	blacklist     *Tools.AccessControlList  // ?
	whitelist     *Tools.AccessControlList  // ?
	ipRateLimiter *Tools.IpRateLimiter      // ? (would be pointless when using a channel communication)

	eventHandler Event.Handler
}

func New[B any, O any](
	name string,
	config *Config.Server,
	acceptHandler Tools.AcceptHandler[Systemge.Connection[B]], // ?
	readHandler Tools.ReadHandler[O, Systemge.Connection[B]], // ?
	eventHandler Event.Handler,
) (*Server[B, O], error) {

	if config == nil {
		return nil, errors.New("config is nil")
	}

	server := &Server[B, O]{
		name:         name,
		config:       config,
		eventHandler: eventHandler,
		instanceId:   Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		acceptHandler:    acceptHandler,
		receptionHandler: readHandler,

		sessionManager: Tools.NewSessionManager(config.SessionManagerConfig, nil, nil),
	}
	return server, nil
}

func (server *Server) GetName() string {
	return server.name
}

func (server *Server) GetStatus() int {
	return server.status
}

func (server *Server) GetInstanceId() string {
	return server.instanceId
}

func (server *Server) GetSessionId() string {
	return server.sessionId
}

func (server *Server) onEvent(event *Event.Event) *Event.Event {
	event.GetContext().Merge(server.GetServerContext())
	if server.eventHandler != nil {
		server.eventHandler(event)
	}
	return event
}
func (server *Server) GetServerContext() Event.Context {
	return Event.Context{
		Event.ServiceType:       Event.SystemgeServer,
		Event.ServiceName:       server.name,
		Event.ServiceStatus:     Status.ToString(server.status),
		Event.Function:          Event.GetCallerFuncName(2),
		Event.ServiceInstanceId: server.instanceId,
		Event.SessionId:         server.sessionId,
	}
}
