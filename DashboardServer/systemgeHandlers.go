package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *Server) onSystemgeConnectHandler(connection SystemgeConnection.SystemgeConnection) error {
	response, err := connection.SyncRequestBlocking(DashboardHelpers.TOPIC_INTRODUCTION, "")
	if err != nil {
		return err
	}

	client, err := DashboardHelpers.UnmarshalIntroduction([]byte(response.GetPayload()))
	if err != nil {
		return err
	}

	connectedClient := newConnectedClient(connection, client)

	server.mutex.Lock()
	server.registerModuleHttpHandlers(connectedClient)
	server.connectedClients[connection.GetName()] = connectedClient
	server.mutex.Unlock()

	server.websocketServer.Multicast(
		// propagate the new clientStatus to all websocket clients on the dashboard-page
		Message.NewAsync(
			"addModule",
			Helpers.JsonMarshal(client),
		),
	)
	return nil
}

func (server *Server) onSystemgeDisconnectHandler(connection SystemgeConnection.SystemgeConnection) {
	server.mutex.Lock()
	if connectedClient, ok := server.connectedClients[connection.GetName()]; ok {
		for websocketClient := range connectedClient.websocketClients {
			server.handleChangePage(websocketClient, Message.NewAsync(DashboardHelpers.TOPIC_CHANGE_PAGE, "/"))
		}
		delete(server.connectedClients, connection.GetName())
		server.unregisterModuleHttpHandlers(connectedClient)
	}
	server.mutex.Unlock()

	server.websocketServer.Multicast(
		// propagate the removed clientStatus to all websocket clients on the dashboard-page
		Message.NewAsync(
			"removeModule",
			connection.GetName(),
		),
	)
	return
}
