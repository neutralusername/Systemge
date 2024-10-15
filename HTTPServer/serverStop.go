package HTTPServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Status"
)

func (server *HTTPServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if server.status != Status.Started {
		return errors.New("http server not started")
	}
	server.status = Status.Pending

	err := server.httpServer.Close()
	if err != nil {
		// something (shoudln't happen)
	}
	server.httpServer = nil
	server.status = Status.Stopped

	return nil
}
