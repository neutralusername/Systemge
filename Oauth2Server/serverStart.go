package Oauth2Server

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *Server) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.Stopped {
		return Event.New("Server is not in stopped state", nil)
	}
	server.status = Status.Pending
	server.sessionRequestChannel = make(chan *oauth2SessionRequest)
	err := server.httpServer.Start()
	if err != nil {
		server.status = Status.Stopped
		close(server.sessionRequestChannel)
		server.sessionRequestChannel = nil
		return err
	}
	go server.handleSessionRequests()
	server.status = Status.Started
	return nil
}
