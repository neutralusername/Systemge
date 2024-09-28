package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *SystemgeServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.Started {
		return Event.New("server is already stopped", nil)
	}
	server.status = Status.Pending
	if server.infoLogger != nil {
		server.infoLogger.Log("stopping server")
	}

	close(server.stopChannel)
	server.listener.Close()

	server.waitGroup.Wait()
	server.stopChannel = nil
	server.listener = nil

	if infoLogger := server.infoLogger; infoLogger != nil {
		infoLogger.Log("server stopped")
	}
	server.status = Status.Stoped
	return nil
}
