package DashboardServer

import "github.com/neutralusername/Systemge/DashboardHelpers"

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

func (server *Server) getDashboardData() *dashboardClient {
	dashboardData := &dashboardClient{
		Name:             "dashboard",
		ClientStatuses:   map[string]int{},
		Commands:         map[string]bool{},
		SystemgeMetrics:  map[string]uint64{},
		WebsocketMetrics: map[string]uint64{},
		HttpMetrics:      map[string]uint64{},
		HeapUsage:        0,
		GoroutineCount:   0,
	}
	for command := range server.dashboardCommandHandlers {
		dashboardData.Commands[command] = true
	}
	for _, connectedClient := range server.connectedClients {
		dashboardData.ClientStatuses[connectedClient.connection.GetName()] = DashboardHelpers.GetStatus(connectedClient.client)
	}
	dashboardData.SystemgeMetrics = server.RetrieveSystemgeMetrics()
	dashboardData.WebsocketMetrics = server.RetrieveWebsocketMetrics()
	dashboardData.HttpMetrics = server.RetrieveHttpMetrics()
	dashboardData.HeapUsage = server.getHeapUsage()
	dashboardData.GoroutineCount = server.getGoroutineCount()
	return dashboardData
}
