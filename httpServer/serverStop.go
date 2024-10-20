package httpServer

import (
	"errors"

	"github.com/neutralusername/systemge/status"
)

func (server *HTTPServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if server.status != status.Started {
		return errors.New("http server not started")
	}
	server.status = status.Pending

	err := server.httpServer.Close()
	if err != nil {
		// something (shoudln't happen)
	}
	server.httpServer = nil
	server.status = status.Stopped

	return nil
}
