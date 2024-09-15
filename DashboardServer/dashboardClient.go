package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Helpers"
)

func (server *Server) GetDashboardClientMetrics() map[string]map[string]uint64 {
	metrics := server.checkResourceUsageMetrics()
	DashboardHelpers.MergeMetrics(metrics, server.websocketServer.GetMetrics())
	DashboardHelpers.MergeMetrics(metrics, server.httpServer.GetMetrics())
	DashboardHelpers.MergeMetrics(metrics, server.systemgeServer.GetMetrics())
	return metrics
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

func (server *Server) checkResourceUsageMetrics() map[string]map[string]uint64 {
	return map[string]map[string]uint64{
		"resource_usage": {
			"heap_usage":      Helpers.HeapUsage(),
			"goroutine_count": uint64(Helpers.GoroutineCount()),
		},
	}
}
