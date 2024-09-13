package DashboardServer

import (
	"runtime"
	"strconv"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (server *Server) pageRequestHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	server.mutex.RLock()
	currentPage := server.websocketClientLocations[websocketClient]
	connectedClient := server.connectedClients[currentPage]
	server.mutex.RUnlock()

	if currentPage == "" {
		return Error.New("No location", nil)
	}

	request, err := Message.Deserialize([]byte(message.GetPayload()), websocketClient.GetId())
	if err != nil {
		return Error.New("Failed to deserialize request", err)
	}

	if currentPage == "/" {
		return server.handleDashboardRequest(websocketClient, request)
	} else {
		if connectedClient == nil {
			// should never happen
			return Error.New("Client not found", nil)
		}
		switch connectedClient.client.(type) {
		case DashboardHelpers.CommandClient:
			return server.handleCommandClientRequest(websocketClient, request, connectedClient)
		case DashboardHelpers.CustomServiceClient:
			return server.handleCustomServiceClientRequest(websocketClient, request, connectedClient)
		case DashboardHelpers.SystemgeConnectionClient:
			return server.handleSystemgeConnectionClientRequest(websocketClient, request, connectedClient)
		default:
			// should never happen
			return Error.New("Unknown client type", nil)
		}
	}
}

func (server *Server) handleDashboardRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_EXECUTE_COMMAND:
		command, err := DashboardHelpers.UnmarshalCommand(request.GetPayload())
		if err != nil {
			return Error.New("Failed to parse command", err)
		}
		commandHandler, _ := server.dashboardCommandHandlers.Get(command.Command)
		if commandHandler == nil {
			return Error.New("Command not found", nil)
		}
		resultPayload, err := commandHandler(command.Args)
		if err != nil {
			return Error.New("Failed to execute command", err)
		}
		websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			resultPayload,
		).Serialize())
		return nil
	case DashboardHelpers.TOPIC_COLLECT_GARBAGE:
		runtime.GC()
		websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			"Garbage collected",
		).Serialize())
		return nil
	case DashboardHelpers.TOPIC_GOROUTINE_COUNT:
		goroutineCount := runtime.NumGoroutine()
		websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			Helpers.IntToString(goroutineCount),
		).Serialize())
		return nil
	case DashboardHelpers.TOPIC_HEAP_USAGE:
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		heapSize := strconv.FormatUint(memStats.HeapSys, 10)
		websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			heapSize,
		).Serialize())
		return nil
	case DashboardHelpers.TOPIC_GET_METRICS:

	case DashboardHelpers.TOPIC_STOP:

	}
}

func (server *Server) getHeapUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.HeapSys
}

func (server *Server) getGoroutineCount() int {
	return runtime.NumGoroutine()
}

func (server *Server) handleCommandClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_EXECUTE_COMMAND:

	}
}

func (server *Server) handleCustomServiceClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_EXECUTE_COMMAND:

	case DashboardHelpers.TOPIC_START:

	case DashboardHelpers.TOPIC_STOP:

	case DashboardHelpers.TOPIC_GET_METRICS:

	case DashboardHelpers.TOPIC_GET_STATUS:

	}
}

func (server *Server) handleSystemgeConnectionClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_EXECUTE_COMMAND:

	case DashboardHelpers.TOPIC_START:

	case DashboardHelpers.TOPIC_GET_METRICS:

	case DashboardHelpers.TOPIC_STOP:

	case DashboardHelpers.TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY:

	case DashboardHelpers.TOPIC_START_PROCESSINGLOOP_CONCURRENTLY:

	case DashboardHelpers.TOPIC_STOP_PROCESSINGLOOP:

	case DashboardHelpers.TOPIC_IS_PROCESSING_LOOP_RUNNING:

	case DashboardHelpers.TOPIC_PROCESS_NEXT_MESSAGE:

	case DashboardHelpers.TOPIC_UNPROCESSED_MESSAGE_COUNT:

	case DashboardHelpers.TOPIC_SYNC_REQUEST:

	case DashboardHelpers.TOPIC_ASYNC_MESSAGE:

	}
}

func (server *Server) handleCommandRequest(websocketClient *WebsocketServer.WebsocketClient, page string, requestPayload string) error {
	command, err := DashboardHelpers.UnmarshalCommand(requestPayload)
	if err != nil {
		return Error.New("Failed to parse command", err)
	}
	switch page {
	case "/":
		commandHandler, _ := server.dashboardCommandHandlers.Get(command.Command)
		if commandHandler == nil {
			return Error.New("Command not found", nil)
		}
		resultPayload, err := commandHandler(command.Args)
		if err != nil {
			return Error.New("Failed to execute command", err)
		}
		websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			resultPayload,
		).Serialize())
	default:
		server.mutex.RLock()
		connectedClient := server.connectedClients[page]
		server.mutex.RUnlock()
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
	}
	return nil
}
func (server *Server) handleStartRequest(websocketClient *WebsocketServer.WebsocketClient, page string) error {
	switch page {
	case "/":
		return Error.New("Cannot start from dashboard", nil)
	default:
		server.mutex.RLock()
		connectedClient := server.connectedClients[page]
		server.mutex.RUnlock()
		if connectedClient == nil {
			return Error.New("Client not found", nil)
		}
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
}
func (server *Server) handleStopRequest(websocketClient *WebsocketServer.WebsocketClient, page string) error {
	switch page {
	case "/":
		return Error.New("Cannot stop from dashboard", nil)
	default:
		server.mutex.RLock()
		connectedClient := server.connectedClients[page]
		server.mutex.RUnlock()
		if connectedClient == nil {
			return Error.New("Client not found", nil)
		}
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
}
func (server *Server) gcHandler(websocketClient *WebsocketServer.WebsocketClient, page string) error {
	switch page {
	case "/":
		runtime.GC()
		return nil
	default:
		return Error.New("Cannot collect garbage from client page", nil)
	}
}
func (server *Server) handleMetricsRequest(websocketClient *WebsocketServer.WebsocketClient, page string, requestPayload string) error {

}
func (server *Server) handleStatusRequest(websocketClient *WebsocketServer.WebsocketClient, page string, requestPayload string) error {

}
func (server *Server) handleHeapUsageRequest(websocketClient *WebsocketServer.WebsocketClient, page string) error {

}
func (server *Server) handleGoroutineCountRequest(websocketClient *WebsocketServer.WebsocketClient, page string) error {

}

func (server *Server) handleChangePage(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	locationBeforeChange := server.websocketClientLocations[websocketClient]
	locationAfterChange := message.GetPayload()

	if locationBeforeChange == locationAfterChange {
		return Error.New("Location is already "+locationAfterChange, nil)
	}

	var page *DashboardHelpers.PageUpdate
	var connectedClient *connectedClient
	switch locationAfterChange {
	case "":
		page = DashboardHelpers.NewPage(
			map[string]interface{}{},
			DashboardHelpers.PAGE_NULL,
		)
	case "/":
		page = DashboardHelpers.NewPage(
			server.getDashboardData(),
			DashboardHelpers.PAGE_DASHBOARD,
		)
		server.dashboardWebsocketClients[websocketClient] = true
	default:
		connectedClient = server.connectedClients[locationAfterChange]
		if connectedClient == nil {
			return Error.New("Client not found", nil)
		}
		page = DashboardHelpers.GetPage(connectedClient.client)
		connectedClient.websocketClients[websocketClient] = true
	}
	server.websocketClientLocations[websocketClient] = locationAfterChange
	switch locationBeforeChange {
	case "":
	case "/":
		delete(server.dashboardWebsocketClients, websocketClient)
	default:
		delete(connectedClient.websocketClients, websocketClient)
	}
	go websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_CHANGE_PAGE,
		page.Marshal(),
	).Serialize())
	return nil
}

func (server *Server) onWebsocketConnectHandler(websocketClient *WebsocketServer.WebsocketClient) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.websocketClientLocations[websocketClient] = ""
	return nil
}
func (server *Server) onWebsocketDisconnectHandler(websocketClient *WebsocketServer.WebsocketClient) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.websocketClientLocations, websocketClient)
}
