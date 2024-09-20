package DashboardServer

import (
	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (server *Server) handleClientCommandRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	_, err := DashboardHelpers.UnmarshalCommand(request.GetPayload())
	if err != nil {
		return Error.New("Failed to parse command", err)
	}
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_COMMAND, request.GetPayload())
	if err != nil {
		return Error.New("Failed to execute command", err)
	}
	server.handleWebsocketResponseMessage(websocketClient, resultPayload)
	return nil
}

func (server *Server) handleClientStartRequest(connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_START, "")
	if err != nil {
		return Error.New("Failed to start client", err)
	}
	err = connectedClient.page.SetCachedStatus(Helpers.StringToInt(resultPayload))
	if err != nil {
		// should never happen
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DashboardHelpers.DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_CLIENTSTATUSES: map[string]int{
						connectedClient.connection.GetName(): Helpers.StringToInt(resultPayload),
					},
				},
				DashboardHelpers.DASHBOARD_CLIENT_NAME,
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

func (server *Server) handleClientStopRequest(connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_STOP, "")
	if err != nil {
		return Error.New("Failed to stop client", err)
	}
	err = connectedClient.page.SetCachedStatus(Helpers.StringToInt(resultPayload))
	if err != nil {
		// should never happen
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(DashboardHelpers.DASHBOARD_CLIENT_NAME),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_MERGE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_CLIENTSTATUSES: map[string]int{
						connectedClient.connection.GetName(): Helpers.StringToInt(resultPayload),
					},
				},
				DashboardHelpers.DASHBOARD_CLIENT_NAME,
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

func (server *Server) handleClientCloseChildRequest(connectedClient *connectedClient, request *Message.Message) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_CLOSE_CHILD, request.GetPayload())
	if err != nil {
		return Error.New("Failed to close child", err)
	}
	systemgeClientChildren := connectedClient.page.GetCachedSystemgeConnectionChildren()
	if systemgeClientChildren == nil {
		// should never happen
		return Error.New("Failed to get systemge connection children", nil)
	}
	delete(systemgeClientChildren, request.GetPayload())
	err = connectedClient.page.SetCachedSystemgeConnectionChildren(systemgeClientChildren)
	if err != nil {
		// should never happen
		return Error.New("Failed to update systemge connection children", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_SYSTEMGE_CONNECTION_CHILDREN: systemgeClientChildren,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) handleClientStartProcessingLoopSequentiallyChildRequest(connectedClient *connectedClient, request *Message.Message) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_SEQUENTIALLY_CHILD, request.GetPayload())
	if err != nil {
		return Error.New("Failed to start processing loop", err)
	}
	systemgeClientChildren := connectedClient.page.GetCachedSystemgeConnectionChildren()
	if systemgeClientChildren == nil {
		// should never happen
		return Error.New("Failed to get systemge connection children", nil)
	}
	systemgeClientChildren[request.GetPayload()].IsMessageHandlingLoopStarted = true
	err = connectedClient.page.SetCachedSystemgeConnectionChildren(systemgeClientChildren)
	if err != nil {
		// should never happen
		return Error.New("Failed to update systemge connection children", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_SYSTEMGE_CONNECTION_CHILDREN: systemgeClientChildren,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) handleClientStartProcessingLoopConcurrentlyChildRequest(connectedClient *connectedClient, request *Message.Message) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_CONCURRENTLY_CHILD, request.GetPayload())
	if err != nil {
		return Error.New("Failed to start processing loop", err)
	}
	systemgeClientChildren := connectedClient.page.GetCachedSystemgeConnectionChildren()
	if systemgeClientChildren == nil {
		// should never happen
		return Error.New("Failed to get systemge connection children", nil)
	}
	systemgeClientChildren[request.GetPayload()].IsMessageHandlingLoopStarted = true
	err = connectedClient.page.SetCachedSystemgeConnectionChildren(systemgeClientChildren)
	if err != nil {
		// should never happen
		return Error.New("Failed to update systemge connection children", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_SYSTEMGE_CONNECTION_CHILDREN: systemgeClientChildren,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) handleClientStopProcessingLoopChildRequest(connectedClient *connectedClient, request *Message.Message) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_STOP_MESSAGE_HANDLING_LOOP_CHILD, request.GetPayload())
	if err != nil {
		return Error.New("Failed to stop processing loop", err)
	}
	systemgeClientChildren := connectedClient.page.GetCachedSystemgeConnectionChildren()
	if systemgeClientChildren == nil {
		// should never happen
		return Error.New("Failed to get systemge connection children", nil)
	}
	systemgeClientChildren[request.GetPayload()].IsMessageHandlingLoopStarted = false
	err = connectedClient.page.SetCachedSystemgeConnectionChildren(systemgeClientChildren)
	if err != nil {
		// should never happen
		return Error.New("Failed to update systemge connection children", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_SYSTEMGE_CONNECTION_CHILDREN: systemgeClientChildren,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) handleClientHandleNextMessageChildRequest(connectedClient *connectedClient, request *Message.Message) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_HANDLE_NEXT_MESSAGE_CHILD, request.GetPayload())
	if err != nil {
		return Error.New("Failed to process next message", err)
	}
	handleNextMessageResult, err := DashboardHelpers.UnmarshalHandleNextMessageResult([]byte(resultPayload))
	if err != nil {
		return Error.New("Failed to parse handle next message result", err)
	}
	systemgeClientChildren := connectedClient.page.GetCachedSystemgeConnectionChildren()
	if systemgeClientChildren == nil {
		// should never happen
		return Error.New("Failed to get systemge connection children", nil)
	}
	systemgeClientChildren[request.GetPayload()].UnhandledMessageCount = handleNextMessageResult.UnhandledMessageCount
	err = connectedClient.page.SetCachedSystemgeConnectionChildren(systemgeClientChildren)
	if err != nil {
		// should never happen
		return Error.New("Failed to update systemge connection children", err)
	}
	server.websocketServer.Multicast(
		server.GetWebsocketClientIdsOnPage(connectedClient.connection.GetName()),
		Message.NewAsync(
			DashboardHelpers.TOPIC_UPDATE_PAGE_REPLACE,
			DashboardHelpers.NewPageUpdate(
				map[string]interface{}{
					DashboardHelpers.CLIENT_FIELD_SYSTEMGE_CONNECTION_CHILDREN: systemgeClientChildren,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) handleClientMultiSyncRequestRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_MULTI_SYNC_REQUEST, request.GetPayload())
	if err != nil {
		return Error.New("Failed to send multi sync request", err)
	}
	server.handleWebsocketResponseMessage(websocketClient, resultPayload)
	return nil
}

func (server *Server) handleClientMultiAsyncMessageRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_MULTI_ASYNC_MESSAGE, request.GetPayload())
	if err != nil {
		return Error.New("Failed to send multi async message", err)
	}
	server.handleWebsocketResponseMessage(websocketClient, "successfully sent multi async message")
	return nil
}

func (server *Server) handleClientStartProcessingLoopSequentiallyRequest(connectedClient *connectedClient) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_SEQUENTIALLY, "")
	if err != nil {
		return Error.New("Failed to start processing loop", err)
	}
	err = connectedClient.page.SetCachedIsProcessingLoopRunning(true)
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
					DashboardHelpers.CLIENT_FIELD_IS_MESSAGE_HANDLING_LOOP_STARTED: true,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) handleClientStartProcessingLoopConcurrentlyRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_CONCURRENTLY, "")
	if err != nil {
		return Error.New("Failed to start processing loop", err)
	}
	err = connectedClient.page.SetCachedIsProcessingLoopRunning(true)
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
					DashboardHelpers.CLIENT_FIELD_IS_MESSAGE_HANDLING_LOOP_STARTED: true,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) handleClientStopProcessingLoopRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_STOP_MESSAGE_HANDLING_LOOP, "")
	if err != nil {
		return Error.New("Failed to stop processing loop", err)
	}
	err = connectedClient.page.SetCachedIsProcessingLoopRunning(false)
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
					DashboardHelpers.CLIENT_FIELD_IS_MESSAGE_HANDLING_LOOP_STARTED: false,
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	return nil
}

func (server *Server) handleClientHandleNextMessageRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_HANDLE_NEXT_MESSAGE, "")
	if err != nil {
		return Error.New("Failed to process next message", err)
	}
	handleNextMessageResult, err := DashboardHelpers.UnmarshalHandleNextMessageResult([]byte(resultPayload))
	if err != nil {
		return Error.New("Failed to parse handle next message result", err)
	}

	err = connectedClient.page.SetCachedUnprocessedMessageCount(handleNextMessageResult.UnhandledMessageCount)
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
					DashboardHelpers.CLIENT_FIELD_UNHANDLED_MESSAGE_COUNT: Helpers.StringToUint32(resultPayload),
				},
				connectedClient.connection.GetName(),
			).Marshal(),
		),
	)
	server.handleWebsocketResponseMessage(websocketClient, resultPayload)
	return nil
}

func (server *Server) handleClientSyncRequestRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {
	resultPayload, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_SYNC_REQUEST, request.GetPayload())
	if err != nil {
		return Error.New("Failed to send sync request", err)
	}
	server.handleWebsocketResponseMessage(websocketClient, resultPayload)
	return nil
}

func (server *Server) handleClientAsyncMessageRequest(websocketClient *WebsocketServer.WebsocketClient, connectedClient *connectedClient, request *Message.Message) error {
	_, err := connectedClient.executeRequest(DashboardHelpers.TOPIC_ASYNC_MESSAGE, request.GetPayload())
	if err != nil {
		return Error.New("Failed to send async message", err)
	}
	server.handleWebsocketResponseMessage(websocketClient, "successfully sent async message")
	return nil
}
