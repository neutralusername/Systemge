package Oauth2Server

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *Server) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.Started {
		return Event.New("Server is not in started state", nil)
	}
	server.status = Status.Pending
	server.httpServer.Stop()
	close(server.sessionRequestChannel)

	server.mutex.Lock()
	for _, session := range server.sessions {
		session.watchdog.Stop()
		session.watchdog = nil
		delete(server.sessions, session.sessionId)
		delete(server.identities, session.identity)
	}
	server.mutex.Unlock()

	server.status = Status.Stopped
	return nil
}
