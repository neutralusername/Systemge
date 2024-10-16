package Server

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Systemge"
	"github.com/neutralusername/Systemge/Tools"
)

// ? == might not make it

type Server[B any, C Systemge.Connection[B]] struct { // this whole struct might be redundant
	instanceId string
	sessionId  string

	status      int
	statusMutex sync.RWMutex
	stopChannel chan bool

	config *Config.Server // ?

	readRoutine   *Tools.Routine
	acceptRoutine *Tools.Routine

	acceptHandler Tools.AcceptHandler[C]  // ?
	readHandler   Tools.ReadHandler[B, C] // ?
	listener      Systemge.Listener[B, C] // ?
}

func New[B any, C Systemge.Connection[B]](
	config *Config.Server, // ?
	listener Systemge.Listener[B, C], // ?
	acceptHandler Tools.AcceptHandler[C], // ?
	readHandler Tools.ReadHandler[B, C], // ?
) (*Server[B, C], error) {

	if config == nil {
		return nil, errors.New("config is nil")
	}

	server := &Server[B, C]{
		config:     config,
		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),

		acceptHandler: acceptHandler,
		readHandler:   readHandler,
	}
	return server, nil
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
