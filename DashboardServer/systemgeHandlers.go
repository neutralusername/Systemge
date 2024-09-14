package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
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
	connectedClient := newConnectedClient(connection, page)

	server.mutex.Lock()
	server.registerModuleHttpHandlers(connectedClient)
	server.connectedClients[connection.GetName()] = connectedClient
	server.mutex.Unlock()

	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPage(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_CLIENTSTATUSES: map[string]int{
						connection.GetName(): page.GetCachedStatus(),
					},
				},
				DashboardHelpers.CLIENT_TYPE_DASHBOARD,
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) onSystemgeDisconnectHandler(connection SystemgeConnection.SystemgeConnection) {
	server.mutex.Lock()
	if connectedClient, ok := server.connectedClients[connection.GetName()]; ok {
		for websocketClient := range connectedClient.websocketClients {
			server.handleChangePage(websocketClient, Message.NewAsync(DashboardHelpers.TOPIC_CHANGE_PAGE, DASHBOARD_CLIENT_NAME))
		}
		delete(server.connectedClients, connection.GetName())
		server.unregisterModuleHttpHandlers(connectedClient)
	}
	server.mutex.Unlock()

	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE, // it would be less awful to have a separate topic for removing keys in an object
			DashboardHelpers.NewPage(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_CLIENTSTATUSES: server.dashboardClient.ClientStatuses,
				},
				DashboardHelpers.CLIENT_TYPE_DASHBOARD,
			).Marshal(),
		),
	)
}
