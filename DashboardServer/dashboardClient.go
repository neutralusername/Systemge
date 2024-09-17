package DashboardServer

import (
	"github.com/neutralusername/Systemge/Metrics"
)

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
