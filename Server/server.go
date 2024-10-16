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

type Server[B any, C Systemge.Connection[B]] struct {
	instanceId string
	sessionId  string
	name       string

	status      int
	statusMutex sync.RWMutex
	stopChannel chan bool

	config *Config.Server

	readRoutine   *Tools.Routine
	acceptRoutine *Tools.Routine

	acceptHandler Tools.AcceptHandler[C]  // ?
	readHandler   Tools.ReadHandler[B, C] // ?
	listener      Systemge.Listener[B, C] // ?

	eventHandler Event.Handler
}

func New[B any, C Systemge.Connection[B]](
	name string,
	config *Config.Server,
	listener Systemge.Listener[B, C], // ?
	acceptHandler Tools.AcceptHandler[C], // ?
	readHandler Tools.ReadHandler[B, C], // ?
	eventHandler Event.Handler, // probably redundant at this place if server only provides the managing of accept and read routines and those handlers are provided by the caller (more appropriate in the handlers) (could have its place if server automatically provides those handlers)
) (*Server[B, C], error) {

	if config == nil {
		return nil, errors.New("config is nil")
	}

	server := &Server[B, C]{
		name:         name,
		config:       config,
		eventHandler: eventHandler,
		instanceId:   Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		acceptHandler: acceptHandler,
		readHandler:   readHandler,
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
