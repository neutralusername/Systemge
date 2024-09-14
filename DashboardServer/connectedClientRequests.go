package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

// handles a command request from a client and sends the result back to the client
func (Server *Server) handleClientCommandRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	_, err := DashboardHelpers.UnmarshalCommand(request.GetPayload())
	if err != nil {
		return Error.New("Failed to parse command", err)
	}
	if connectedClient == nil {
		return Error.New("Client not found", nil)
	}
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_COMMAND, request.GetPayload())
	if err != nil {
		return Error.New("Failed to execute command", err)
	}
	websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
		resultPayload,
	).Serialize())
	return nil
}

// starts the client and updates the status to everyone who should be informed
func (server *Server) handleClientStartRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	newStatus, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_START, "")
	if err != nil {
		return Error.New("Failed to start client", err)
	}
	err = DashboardHelpers.SetStatus(connectedClient.client, Helpers.StringToInt(newStatus))
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

// stops the client and updates the status to everyone who should be informed
func (server *Server) handleClientStopRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	newStatus, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_STOP, "")
	if err != nil {
		return Error.New("Failed to stop client", err)
	}
	err = DashboardHelpers.SetStatus(connectedClient.client, Helpers.StringToInt(newStatus))
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

// starts the clients processing loop sequentially and updates the status to everyone who should be informed
func (server *Server) handleClientStartProcessingLoopSequentiallyRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {

}

// starts the clients processing loop concurrently and updates the status to everyone who should be informed
func (server *Server) handleClientStartProcessingLoopConcurrentlyRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {

}

// stops the clients processing loop and updates the status to everyone who should be informed
func (server *Server) handleClientStopProcessingLoopRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {

}

// processes the next message in the clients processing loop and sends the result back to the client - also receives an update on unprocessed messages count which is propagated to everyone who should be informed
func (server *Server) handleClientProcessNextMessageRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {

}

func (server *Server) handleClientSyncRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {

}

func (server *Server) handleClientAsyncMessageRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {

}
