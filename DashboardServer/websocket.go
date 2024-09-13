package DashboardServer

import (
	"runtime"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

/* switch request.GetTopic() {
case DashboardHelpers.REQUEST_COMMAND:
	err = server.handleCommandRequest(websocketClient, currentPage, request.GetPayload())
	if err != nil {
		return Error.New("Failed to handle command request", err)
	}
case DashboardHelpers.REQUEST_START:
	err := server.handleStartRequest(websocketClient, currentPage)
	if err != nil {
		return Error.New("Failed to handle start request", err)
	}
case DashboardHelpers.REQUEST_STOP:
	err := server.handleStopRequest(websocketClient, currentPage)
	if err != nil {
		return Error.New("Failed to handle stop request", err)
	}
case DashboardHelpers.REQUEST_COLLECTGARBAGE:
	err := server.gcHandler(websocketClient, currentPage)
	if err != nil {
		return Error.New("Failed to handle garbage collection request", err)
	}
case DashboardHelpers.REQUEST_METRICS:
	err := server.handleMetricsRequest(websocketClient, currentPage, request.GetPayload())
	if err != nil {
		return Error.New("Failed to handle metrics request", err)
	}
case DashboardHelpers.REQUEST_STATUS:
	err := server.handleStatusRequest(websocketClient, currentPage, request.GetPayload())
	if err != nil {
		return Error.New("Failed to handle status request", err)
	}
case DashboardHelpers.REQUEST_HEAPUSAGE:
	err := server.handleHeapUsageRequest(websocketClient, currentPage)
	if err != nil {
		return Error.New("Failed to handle heap usage request", err)
	}
case DashboardHelpers.REQUEST_GOROUTINECOUNT:
	err := server.handleGoroutineCountRequest(websocketClient, currentPage)
	if err != nil {
		return Error.New("Failed to handle goroutine count request", err)
	}
default:
	return Error.New("Unknown request topic", nil)
}
return nil */

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
	case DashboardHelpers.REQUEST_COMMAND:

	case DashboardHelpers.REQUEST_COLLECTGARBAGE:

	case DashboardHelpers.REQUEST_GOROUTINECOUNT:

	case DashboardHelpers.REQUEST_METRICS:

	case DashboardHelpers.REQUEST_HEAPUSAGE:

	}
}

func (server *Server) handleCommandClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.REQUEST_COMMAND:

	}
}

func (server *Server) handleCustomServiceClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.REQUEST_COMMAND:

	case DashboardHelpers.REQUEST_START:

	case DashboardHelpers.REQUEST_STOP:

	case DashboardHelpers.REQUEST_METRICS:

	case DashboardHelpers.REQUEST_STATUS:

	}
}

func (server *Server) handleSystemgeConnectionClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.REQUEST_COMMAND:

	case DashboardHelpers.REQUEST_STATUS:

	case DashboardHelpers.REQUEST_METRICS:

	case DashboardHelpers.REQUEST_STOP:

	case DashboardHelpers.REQUEST_START_PROCESSINGLOOP_SEQUENTIALLY:

	case DashboardHelpers.REQUEST_START_PROCESSINGLOOP_CONCURRENTLY:

	case DashboardHelpers.REQUEST_STOP_PROCESSINGLOOP:

	case DashboardHelpers.REQUEST_IS_PROCESSING_LOOP_RUNNING:

	case DashboardHelpers.REQUEST_PROCESS_NEXT_MESSAGE:

	case DashboardHelpers.REQUEST_UNPROCESSED_MESSAGES_COUNT:

	case DashboardHelpers.REQUEST_SYNC_REQUEST:

	case DashboardHelpers.REQUEST_ASYNC_MESSAGE:

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
func (server *Server) getDashboardData() map[string]interface{} {
	dashboardData := map[string]interface{}{}
	clientStatus := map[string]int{}
	dashboardData["clientStatuses"] = clientStatus
	for _, connectedClient := range server.connectedClients {
		clientStatus[connectedClient.connection.GetName()] = DashboardHelpers.GetStatus(connectedClient.client)
	}
	dashboardCommandsSlice := []string{}
	dashboardData["commands"] = dashboardCommandsSlice
	for command := range server.dashboardCommandHandlers {
		dashboardData["commands"] = append(dashboardCommandsSlice, command)
	}
	dashboardData["systemgeMetrics"] = server.RetrieveSystemgeMetrics()
	dashboardData["websocketMetrics"] = server.RetrieveWebsocketMetrics()
	dashboardData["httpMetrics"] = server.RetrieveHttpMetrics()
	return dashboardData
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
