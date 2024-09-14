package DashboardServer

import (
	"runtime"

	"github.com/neutralusername/Systemge/DashboardHelpers"
)

const (
	DASHBOARD_CLIENT_FIELD_NAME            = "name"
	DASHBOARD_CLIENT_FIELD_CLIENT_STATUSES = "clientStatuses"
	DASHBOARD_CLIENT_FIELD_COMMANDS        = "commands"
	DASHBOARD_CLIENT_FIELD_HEAP_USAGE      = "heapUsage"
	DASHBOARD_CLIENT_FIELD_GOROUTINE_COUNT = "goroutineCount"
	DASHBOARD_CLIENT_FIELD_METRICS         = "metrics"

	DASHBOARD_CLIENT_FIELD_SYSTEMGE_METRICS  = "systemgeMetrics"
	DASHBOARD_CLIENT_FIELD_WEBSOCKET_METRICS = "websocketMetrics"
	DASHBOARD_CLIENT_FIELD_HTTP_METRICS      = "httpMetrics"
)

type dashboardClient struct {
	Name           string                       `json:"name"`
	ClientStatuses map[string]int               `json:"clientStatuses"`
	Commands       map[string]bool              `json:"commands"`
	Metrics        map[string]map[string]uint64 `json:"metrics"`
	HeapUsage      uint64                       `json:"heapUsage"`
	GoroutineCount int                          `json:"goroutineCount"`
}

func (server *Server) getDashboarClient() *dashboardClient {
	dashboardData := &dashboardClient{
		Name:           "dashboard",
		ClientStatuses: server.getClientStatuses(),
		Commands:       server.getDashboardClientCommands(),
		Metrics:        server.getDashboardClientMetrics(),
		HeapUsage:      server.getHeapUsage(),
		GoroutineCount: server.getGoroutineCount(),
	}
	return dashboardData
}

func (server *Server) getDashboardClientCommands() map[string]bool {
	commands := map[string]bool{}
	for command := range server.dashboardCommandHandlers {
		commands[command] = true
	}
	return commands
}

func (server *Server) getClientStatuses() map[string]int {
	clientStatuses := map[string]int{}
	for _, connectedClient := range server.connectedClients {
		clientStatuses[connectedClient.connection.GetName()] = DashboardHelpers.GetStatus(connectedClient.client)
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

// cache this somewhere or else it won't work properly with >1 websocket client
func (server *Server) getDashboardClientMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		DASHBOARD_CLIENT_FIELD_SYSTEMGE_METRICS:  server.RetrieveSystemgeMetrics(),
		DASHBOARD_CLIENT_FIELD_WEBSOCKET_METRICS: server.RetrieveWebsocketMetrics(),
		DASHBOARD_CLIENT_FIELD_HTTP_METRICS:      server.RetrieveHttpMetrics(),
	}
}
