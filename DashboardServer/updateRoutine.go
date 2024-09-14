package DashboardServer

import (
	"time"

	"github.com/neutralusername/Systemge/Status"
)

func (server *Server) updateRoutine() {
	defer server.waitGroup.Done()
	for server.status == Status.STARTED {
		// update on all connected clients their statuses that may change over time and add the newest metrics to the cache/check entries in cache (also on dashboard)
		time.Sleep(time.Duration(server.config.UpdateIntervalMs) * time.Millisecond)
	}
}
