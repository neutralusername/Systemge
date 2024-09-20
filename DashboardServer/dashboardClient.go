package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Metrics"
)

func (server *Server) addMetricsToDashboardClient(metrics map[string]*Metrics.Metrics) {
	for metricsType, metrics := range metrics {
		if server.dashboardClient.Metrics[metricsType] == nil {
			server.dashboardClient.Metrics[metricsType] = []*Metrics.Metrics{}
		}
		server.dashboardClient.Metrics[metricsType] = append(server.dashboardClient.Metrics[metricsType], metrics)
		if server.config.MaxEntriesPerMetrics > 0 && len(server.dashboardClient.Metrics[metricsType]) > server.config.MaxEntriesPerMetrics {
			server.dashboardClient.Metrics[metricsType] = server.dashboardClient.Metrics[metricsType][1:]
		}
	}
}

func (server *Server) updateDashboardClientMetrics() error {
	newMetrics := server.GetMetrics()
	server.addMetricsToDashboardClient(newMetrics)
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DashboardHelpers.DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_METRICS: newMetrics,
				},
				DashboardHelpers.DASHBOARD_CLIENT_NAME,
			).Marshal(),
		),
	)
	return nil
}
