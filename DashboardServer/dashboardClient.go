package DashboardServer

import (
	"runtime"

	"github.com/neutralusername/Systemge/DashboardHelpers"
)

const (
	DASHBOARD_CLIENT_NAME           = "/"
	DASHBOARD_METRICSTYPE_SYSTEMGE  = "systemgeMetrics"
	DASHBOARD_METRICSTYPE_WEBSOCKET = "websocketMetrics"
	DASHBOARD_METRICSTYPE_HTTP      = "httpMetrics"
)

type dashboardClient struct {
	Name           string                                                 `json:"name"`
	ClientStatuses map[string]int                                         `json:"clientStatuses"`
	Commands       map[string]bool                                        `json:"commands"`
	Metrics        map[string]map[string][]*DashboardHelpers.MetricsEntry `json:"metrics"`
	HeapUsage      uint64                                                 `json:"heapUsage"`
	GoroutineCount int                                                    `json:"goroutineCount"`
}

// EXPECTS MUTEX TO BE LOCKED RN
func (server *Server) getClientStatuses() map[string]int {
	clientStatuses := map[string]int{}
	for _, connectedClient := range server.connectedClients {
		clientStatuses[connectedClient.connection.GetName()] = connectedClient.page.GetCachedStatus()
	}
	return clientStatuses
}

func (server *Server) getHeapUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.HeapSys
}

func (server *Server) getGoroutineCount() int {
	return runtime.NumGoroutine()
}

func (server *Server) retrieveDashboardClientMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		DASHBOARD_METRICSTYPE_SYSTEMGE:  server.RetrieveSystemgeMetrics(),
		DASHBOARD_METRICSTYPE_WEBSOCKET: server.RetrieveWebsocketMetrics(),
		DASHBOARD_METRICSTYPE_HTTP:      server.RetrieveHttpMetrics(),
	}
}

func (server *Server) GetSystemgeMetrics() map[string]uint64 {
	return server.systemgeServer.GetMetrics()
}
func (server *Server) RetrieveSystemgeMetrics() map[string]uint64 {
	return server.systemgeServer.RetrieveMetrics()
}

func (server *Server) GetWebsocketMetrics() map[string]uint64 {
	return server.websocketServer.GetMetrics()
}
func (server *Server) RetrieveWebsocketMetrics() map[string]uint64 {
	return server.websocketServer.RetrieveMetrics()
}

func (server *Server) GetHttpMetrics() map[string]uint64 {
	return server.httpServer.GetMetrics()
}
func (server *Server) RetrieveHttpMetrics() map[string]uint64 {
	return server.httpServer.RetrieveMetrics()
}
