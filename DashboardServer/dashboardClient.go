package DashboardServer

import (
	"time"

	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Metrics"
)

func (server *Server) GetDashboardClientMetrics() map[string]*Metrics.Metrics {
	metrics := server.checkResourceUsageMetrics()
	Metrics.Merge(metrics, server.websocketServer.GetMetrics())
	Metrics.Merge(metrics, server.httpServer.GetMetrics())
	Metrics.Merge(metrics, server.systemgeServer.GetMetrics())
	return metrics
}

func (server *Server) addMetricsToDashboardClient(metrics map[string]*Metrics.Metrics) {
	for metricsType, metrics := range metrics {
		if server.dashboardClient.Metrics[metricsType] == nil {
			server.dashboardClient.Metrics[metricsType] = []*Metrics.Metrics{}
		}
		server.dashboardClient.Metrics[metricsType] = append(server.dashboardClient.Metrics[metricsType], metrics)
		if server.config.MaxMetricEntries > 0 && len(server.dashboardClient.Metrics[metricsType]) > server.config.MaxMetricEntries {
			server.dashboardClient.Metrics[metricsType] = server.dashboardClient.Metrics[metricsType][1:]
		}
	}
}

func (server *Server) checkResourceUsageMetrics() map[string]*Metrics.Metrics {
	return map[string]*Metrics.Metrics{
		"resource_usage": {
			KeyValuePairs: map[string]uint64{
				"heap_usage":      Helpers.HeapUsage(),
				"goroutine_count": uint64(Helpers.GoroutineCount()),
			},
			Time: time.Now(),
		},
	}
}
