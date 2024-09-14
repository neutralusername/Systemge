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
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_START, "")
	if err != nil {
		return Error.New("Failed to start client", err)
	}
	err = DashboardHelpers.SetStatus(connectedClient.client, Helpers.StringToInt(resultPayload))
	if err != nil {
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage("/"),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_APPEND,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					"clientStatuses": map[string]int{
						connectedClient.connection.GetName(): Helpers.StringToInt(resultPayload),
					},
				},
				"/",
			).Marshal(),
		),
	)
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_APPEND,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					"status": resultPayload,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

// stops the client and updates the status to everyone who should be informed
func (server *Server) handleClientStopRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_STOP, "")
	if err != nil {
		return Error.New("Failed to stop client", err)
	}
	err = DashboardHelpers.SetStatus(connectedClient.client, Helpers.StringToInt(resultPayload))
	if err != nil {
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage("/"),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_APPEND,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					"clientStatuses": map[string]int{
						connectedClient.connection.GetName(): Helpers.StringToInt(resultPayload),
					},
				},
				"/",
			).Marshal(),
		),
	)
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_APPEND,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					"status": resultPayload,
				},
				connectedClient.connection.GetName(),
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

// handles a outgoing message request from a client and sends the result back to the client
func (server *Server) handleClientSyncRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {

}

// handles a incoming message request from a client and sends the result back to the client
func (server *Server) handleClientAsyncMessageRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {

}
