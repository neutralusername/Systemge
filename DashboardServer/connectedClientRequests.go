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
	err = DashboardHelpers.SetCachedStatus(connectedClient.client, Helpers.StringToInt(resultPayload))
	if err != nil {
		// should never happen
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_APPEND,
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
			DashboardHelpers.TOPIC_UPDATE_PAGE_APPEND,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_STATUS: resultPayload,
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
	err = DashboardHelpers.SetCachedStatus(connectedClient.client, Helpers.StringToInt(resultPayload))
	if err != nil {
		// should never happen
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_APPEND,
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
			DashboardHelpers.TOPIC_UPDATE_PAGE_APPEND,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_STATUS: resultPayload,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

// starts the clients processing loop sequentially and updates the status to everyone who should be informed
func (server *Server) handleClientStartProcessingLoopSequentiallyRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY, "")
	if err != nil {
		return Error.New("Failed to start processing loop", err)
	}
	err = DashboardHelpers.SetCachedIsProcessingLoopRunning(connectedClient.client, true)
	if err != nil {
		// should never happen
		return Error.New("Failed to update processing loop status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_IS_PROCESSING_LOOP_RUNNING: true,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

// starts the clients processing loop concurrently and updates the status to everyone who should be informed
func (server *Server) handleClientStartProcessingLoopConcurrentlyRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_START_PROCESSINGLOOP_CONCURRENTLY, "")
	if err != nil {
		return Error.New("Failed to start processing loop", err)
	}
	err = DashboardHelpers.SetCachedIsProcessingLoopRunning(connectedClient.client, true)
	if err != nil {
		// should never happen
		return Error.New("Failed to update processing loop status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_IS_PROCESSING_LOOP_RUNNING: true,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

// stops the clients processing loop and updates the status to everyone who should be informed
func (server *Server) handleClientStopProcessingLoopRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_STOP_PROCESSINGLOOP, "")
	if err != nil {
		return Error.New("Failed to stop processing loop", err)
	}
	err = DashboardHelpers.SetCachedIsProcessingLoopRunning(connectedClient.client, false)
	if err != nil {
		// should never happen
		return Error.New("Failed to update processing loop status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_IS_PROCESSING_LOOP_RUNNING: false,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

// processes the next message in the clients processing loop and sends the result back to the client - also receives an update on unprocessed messages count which is propagated to everyone who should be informed
func (server *Server) handleClientProcessNextMessageRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_PROCESS_NEXT_MESSAGE, "")
	if err != nil {
		return Error.New("Failed to process next message", err)
	}
	err = DashboardHelpers.SetCachedUnprocessedMessageCount(connectedClient.client, Helpers.StringToUint32(resultPayload))
	if err != nil {
		// should never happen
		return Error.New("Failed to update processing loop status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					"unprocessedMessagesCount": resultPayload,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

// handles a outgoing message request from a client and sends the result back to the client
func (server *Server) handleClientSyncRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_SYNC_REQUEST, request.GetPayload())
	if err != nil {
		return Error.New("Failed to send sync request", err)
	}
	websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
		resultPayload,
	).Serialize())
	return nil
}

// handles a incoming message request from a client and sends the result back to the client
func (server *Server) handleClientAsyncMessageRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_ASYNC_MESSAGE, request.GetPayload())
	if err != nil {
		return Error.New("Failed to send async message", err)
	}
	websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
		"successfully sent async message",
	).Serialize())
	return nil
}
