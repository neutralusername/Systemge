package DashboardServer

import (
	"time"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
)

func (server *Server) updateRoutine() {
	defer server.waitGroup.Done()
	for server.status == Status.STARTED {
		// update on all connected clients their statuses that may change over time and add the newest metrics to the cache/check number of entries in cache (also on dashboard)
		for _, connectedClient := range server.connectedClients {
			switch connectedClient.page.Type {
			case DashboardHelpers.CLIENT_TYPE_COMMAND:
				server.updateConnectedClientStatus(connectedClient)
				server.updateConnectedClientMetrics(connectedClient)
			case DashboardHelpers.CLIENT_TYPE_CUSTOMSERVICE:

			case DashboardHelpers.CLIENT_TYPE_SYSTEMGECONNECTION:

			}
		}
		dashboardClient := server.dashboardClient

		time.Sleep(time.Duration(server.config.UpdateIntervalMs) * time.Millisecond)
	}
}

func (server *Server) updateConnectedClientStatus(connectedClient *connectedClient) {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_GET_STATUS, "")
	if err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log(Error.New("Failed to get status from connected client in updateRoutine \""+connectedClient.connection.GetName()+"\"", err).Error())
		}
	} else {
		connectedClient.page.SetCachedStatus(Helpers.StringToInt(resultPayload))
		server.websocketServer.Multicast(
			server.GetWebsocketClientIdsOnPage(DASHBOARD_CLIENT_NAME),
			Message.NewAsync(
				DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
				DashboardHelpers.NewPageUpdate(
					map[string]interface{}{
						DashboardHelpers.CLIENT_FIELD_CLIENTSTATUSES: map[string]int{
							connectedClient.connection.GetName(): Helpers.StringToInt(resultPayload),
						},
					},
					DASHBOARD_CLIENT_NAME,
				).Marshal(),
			),
		)
		server.websocketServer.Multicast(
			server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
			Message.NewAsync(
				DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
				DashboardHelpers.NewPageUpdate(
					map[string]interface{}{
						DashboardHelpers.CLIENT_FIELD_STATUS: Helpers.StringToInt(resultPayload),
					},
					connectedClient.connection.GetName(),
				).Marshal(),
			),
		)
	}
}

func (server *Server) updateConnectedClientMetrics(connectedClient *connectedClient) {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_GET_METRICS, "")
	if err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log(Error.New("Failed to get metrics from connected client in updateRoutine \""+connectedClient.connection.GetName()+"\"", err).Error())
		}
	} else {
		metrics := DashboardHelpers.UnmarshalMetrics(resultPayload)
		for metricsName, metricsKeyValuePairs := range metrics {
			for key, value := range metricsKeyValuePairs {
				connectedClient.page.AddCachedMetricsEntry(metricsName, key, value, server.config.MaxMetricsCacheValues)
			}
		}
		server.websocketServer.Multicast(
			server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
			Message.NewAsync(
				DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
				DashboardHelpers.NewPageUpdate(
					map[string]interface{}{
						DashboardHelpers.CLIENT_FIELD_METRICS: metrics,
					},
					connectedClient.connection.GetName(),
				).Marshal(),
			),
		)
	}
}
