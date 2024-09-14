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
		for _, connectedClient := range server.connectedClients {
			switch connectedClient.page.Type {
			case DashboardHelpers.CLIENT_TYPE_COMMAND:
				if err := server.updateConnectedClientMetrics(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update metrics for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
			case DashboardHelpers.CLIENT_TYPE_CUSTOMSERVICE:
				if err := server.updateConnectedClientStatus(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update status for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientMetrics(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update metrics for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
			case DashboardHelpers.CLIENT_TYPE_SYSTEMGECONNECTION:
				if err := server.updateConnectedClientStatus(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update status for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientMetrics(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update metrics for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientUnprocessedMessageCount(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update unprocessed message count for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
				if err := server.updateConnectedClientIsProcessingLoopRunning(connectedClient); err != nil {
					if server.errorLogger != nil {
						server.errorLogger.Log(Error.New("Failed to update is processing loop running for \""+connectedClient.connection.GetName()+"\" in updateRoutine", err).Error())
					}
				}
			}
		}
		if err := server.updateDashboardClientMetrics(); err != nil {
			if server.errorLogger != nil {
				server.errorLogger.Log(Error.New("Failed to update metrics for dashboard client in updateRoutine", err).Error())
			}
		}
		time.Sleep(time.Duration(server.config.UpdateIntervalMs) * time.Millisecond)
	}
}

func (server *Server) updateConnectedClientStatus(connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_GET_STATUS, "")
	if err != nil {
		return Error.New("Failed to execute get status request", err)
	}
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
	return nil
}

func (server *Server) updateConnectedClientMetrics(connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_GET_METRICS, "")
	if err != nil {
		if server.errorLogger != nil {
			server.errorLogger.Log(Error.New("Failed to execute get metrics request", err).Error())
		}
	}
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
	return nil
}

func (server *Server) updateConnectedClientUnprocessedMessageCount(connectedClient *connectedClient) error {

}

func (server *Server) updateConnectedClientIsProcessingLoopRunning(connectedClient *connectedClient) error {

}

func (server *Server) updateDashboardClientMetrics() error {

}
