package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *Server) onSystemgeConnectHandler(connection SystemgeConnection.SystemgeConnection) error {
	response, err := connection.SyncRequestBlocking(DashboardHelpers.TOPIC_INTRODUCTION, "")
	if err != nil {
		return err
	}

	page, err := DashboardHelpers.UnmarshalPage([]byte(response.GetPayload()))
	if err != nil {
		return err
	}
	metrics := page.GetCachedMetrics()
	if server.config.MaxMetricTypes > 0 && len(metrics) > server.config.MaxMetricTypes {
		return Error.New("Too many metric types", nil)
	}
	for metricType, metricMap := range metrics {
		if server.config.MaxMetricsPerType > 0 && len(metricMap) > server.config.MaxMetricsPerType {
			return Error.New("Too many metrics of type "+metricType, nil)
		}
		for metricName, metricEntries := range metricMap {
			if server.config.MaxMetricEntries > 0 && len(metricEntries) > server.config.MaxMetricEntries {
				return Error.New("Too many metric entries of type "+metricType+" and name "+metricName, nil)
			}
		}
	}
	commands := page.GetCachedCommands()
	if server.config.MaxCommands > 0 && len(commands) > server.config.MaxCommands {
		return Error.New("Too many commands", nil)
	}

	connectedClient := newConnectedClient(connection, page)

	server.mutex.Lock()
	server.registerModuleHttpHandlers(connectedClient)
	server.connectedClients[connection.GetName()] = connectedClient
	server.dashboardClient.ClientStatuses[connection.GetName()] = page.GetCachedStatus()
	server.mutex.Unlock()

	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DashboardHelpers.DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_CLIENTSTATUSES: map[string]int{
						connection.GetName(): page.GetCachedStatus(),
					},
				},
				DashboardHelpers.DASHBOARD_CLIENT_NAME,
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) onSystemgeDisconnectHandler(connection SystemgeConnection.SystemgeConnection) {
	server.mutex.Lock()
	if connectedClient, ok := server.connectedClients[connection.GetName()]; ok {
		for websocketClient := range connectedClient.websocketClients {
			server.handleChangePage(websocketClient, Message.NewAsync(DashboardHelpers.TOPIC_CHANGE_PAGE, DashboardHelpers.DASHBOARD_CLIENT_NAME))
		}
		delete(server.connectedClients, connection.GetName())
		delete(server.dashboardClient.ClientStatuses, connection.GetName())
		server.unregisterModuleHttpHandlers(connectedClient)
	}
	server.mutex.Unlock()

	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DashboardHelpers.DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE, // it would be less awful to have a separate topic for removing keys in an object
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_CLIENTSTATUSES: server.dashboardClient.ClientStatuses,
				},
				DashboardHelpers.DASHBOARD_CLIENT_NAME,
			).Marshal(),
		),
	)
}
