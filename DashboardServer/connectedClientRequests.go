package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (Server *Server) handleClientCommandRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	command, err := DashboardHelpers.UnmarshalCommand(request.GetPayload())
	if err != nil {
		return Error.New("Failed to parse command", err)
	}
	if connectedClient == nil {
		return Error.New("Client not found", nil)
	}
	resultPayload, err := connectedClient.executeCommand(command.Command, command.Args)
	if err != nil {
		return Error.New("Failed to execute command", err)
	}
	websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
		resultPayload,
	).Serialize())
	return nil
}

func (server *Server) handleClientStartRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	newStatus, err := connectedClient.executeStart()
	if err != nil {
		return Error.New("Failed to start client", err)
	}
	err = DashboardHelpers.SetStatus(connectedClient.client, newStatus)
	if err != nil {
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Multicast(
		// propagate the new clientStatus to all websocket clients on the dashboard-page and on this client-page
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE,
			DashboardHelpers.NewPage(
				map[string]interface{}{
					"status": newStatus,
				},
				DashboardHelpers.GetPageType(connectedClient.client),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) handleClientStopRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	newStatus, err := connectedClient.executeStop()
	if err != nil {
		return Error.New("Failed to stop client", err)
	}
	err = DashboardHelpers.SetStatus(connectedClient.client, newStatus)
	if err != nil {
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Multicast(
		// propagate the new clientStatus to all websocket clients on the dashboard-page and on this client-page
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE,
			DashboardHelpers.NewPage(
				map[string]interface{}{
					"status": newStatus,
				},
				DashboardHelpers.GetPageType(connectedClient.client),
			).Marshal(),
		),
	)
	return nil
}
