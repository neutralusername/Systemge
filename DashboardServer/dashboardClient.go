package DashboardServer

import "github.com/neutralusername/Systemge/DashboardHelpers"

const (
	DASHBOARD_CLIENT_FIELD_NAME              = "name"
	DASHBOARD_CLIENT_FIELD_CLIENT_STATUSES   = "clientStatuses"
	DASHBOARD_CLIENT_FIELD_COMMANDS          = "commands"
	DASHBOARD_CLIENT_FIELD_SYSTEMGE_METRICS  = "systemgeMetrics"
	DASHBOARD_CLIENT_FIELD_WEBSOCKET_METRICS = "websocketMetrics"
	DASHBOARD_CLIENT_FIELD_HTTP_METRICS      = "httpMetrics"
	DASHBOARD_CLIENT_FIELD_HEAP_USAGE        = "heapUsage"
	DASHBOARD_CLIENT_FIELD_GOROUTINE_COUNT   = "goroutineCount"
)

type dashboardClient struct {
	Name             string            `json:"name"`
	ClientStatuses   map[string]int    `json:"clientStatuses"`
	Commands         map[string]bool   `json:"commands"`
	SystemgeMetrics  map[string]uint64 `json:"systemgeMetrics"`
	WebsocketMetrics map[string]uint64 `json:"websocketMetrics"`
	HttpMetrics      map[string]uint64 `json:"httpMetrics"`
	HeapUsage        uint64            `json:"heapUsage"`
	GoroutineCount   int               `json:"goroutineCount"`
}

func (server *Server) getDashboarClient() *dashboardClient {
	dashboardData := &dashboardClient{
		Name:             "dashboard",
		ClientStatuses:   server.getClientStatuses(),
		Commands:         server.getDashboardClientCommands(),
		SystemgeMetrics:  server.RetrieveSystemgeMetrics(),
		WebsocketMetrics: server.RetrieveWebsocketMetrics(),
		HttpMetrics:      server.RetrieveHttpMetrics(),
		HeapUsage:        server.getHeapUsage(),
		GoroutineCount:   server.getGoroutineCount(),
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
