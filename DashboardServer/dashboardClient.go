package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Helpers"
)

func (server *Server) GetDashboardClientMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		DashboardHelpers.DASHBOARD_METRICSTYPE_SYSTEMGE:      server.GetSystemgeMetrics(),
		DashboardHelpers.DASHBOARD_METRICSTYPE_WEBSOCKET:     server.GetWebsocketMetrics(),
		DashboardHelpers.DASHBOARD_METRICSTYPE_HTTP:          server.GetHttpMetrics(),
		DashboardHelpers.DASHBOARD_METRICSTYPE_RESOURCEUSAGE: server.CheckResourceUsageMetrics(),
	}
}

func (server *Server) addMetricsToDashboardClient(metrics map[string]map[string]*DashboardHelpers.MetricsEntry) {
	for metricsType, metricsKeyValuePairs := range metrics {
		if server.dashboardClient.Metrics[metricsType] == nil {
			server.dashboardClient.Metrics[metricsType] = map[string][]*DashboardHelpers.MetricsEntry{}
		}
		for key, value := range metricsKeyValuePairs {
			server.dashboardClient.Metrics[metricsType][key] = append(server.dashboardClient.Metrics[metricsType][key], value)
			if len(server.dashboardClient.Metrics[metricsType][key]) > server.config.MaxMetricEntries {
				server.dashboardClient.Metrics[metricsType][key] = server.dashboardClient.Metrics[metricsType][key][1:]
			}
		}
	}
}

func (server *Server) CheckResourceUsageMetrics() map[string]uint64 {
	return map[string]uint64{
		"heap_usage":      Helpers.HeapUsage(),
		"goroutine_count": uint64(Helpers.GoroutineCount()),
	}
}

func (server *Server) CheckSystemgeMetrics() map[string]uint64 {
	metrics := server.systemgeServer.CheckMetrics()
	for _, value := range metrics {
		return value
	}
	return nil
}
func (server *Server) GetSystemgeMetrics() map[string]uint64 {
	metrics := server.systemgeServer.GetMetrics()
	for _, value := range metrics {
		return value
	}
	return nil
}

func (server *Server) CheckWebsocketMetrics() map[string]uint64 {
	metrics := server.websocketServer.CheckMetrics()
	for _, value := range metrics {
		return value
	}
	return nil
}
func (server *Server) GetWebsocketMetrics() map[string]uint64 {
	metrics := server.websocketServer.GetMetrics()
	for _, value := range metrics {
		return value
	}
	return nil
}

func (server *Server) CheckHttpMetrics() map[string]uint64 {
	metrics := server.httpServer.CheckMetrics()
	for _, value := range metrics {
		return value
	}
	return nil
}
func (server *Server) GetHttpMetrics() map[string]uint64 {
	metrics := server.httpServer.GetMetrics()
	for _, value := range metrics {
		return value
	}
	return nil
}
