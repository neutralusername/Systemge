package DashboardServer

import (
	"time"

	"github.com/neutralusername/Systemge/Status"
)

func (server *Server) updateRoutine() {
	defer server.waitGroup.Done()
	for server.status == Status.STARTED {

		time.Sleep(time.Duration(server.config.UpdateIntervalMs) * time.Millisecond)
	}
}
