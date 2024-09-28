package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/TcpSystemgeListener"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *SystemgeServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)

	if server.status != Status.Stoped {
		return Event.New("server is already started", nil)
	}
	server.status = Status.Pending
	if server.infoLogger != nil {
		server.infoLogger.Log("starting server")
	}
	listener, err := TcpSystemgeListener.New(server.name, server.config.TcpSystemgeListenerConfig, server.whitelist, server.blacklist, server.eventHandler)
	if err != nil {
		server.status = Status.Stoped
		return Event.New("failed to create listener", err)
	}
	server.listener = listener
	server.stopChannel = make(chan bool)
	go server.handleConnections(server.stopChannel)

	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log("server started")
	}
	server.status = Status.Started
	return nil
}
