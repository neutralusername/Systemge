package DashboardServer

import (
	"runtime"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (server *Server) pageRequestHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	server.mutex.RLock()
	currentPage := server.websocketClientLocations[websocketClient]
	server.mutex.RUnlock()

	if currentPage == "" {
		return Error.New("No location", nil)
	}

	request, err := Message.Deserialize([]byte(message.GetPayload()), websocketClient.GetId())
	if err != nil {
		return Error.New("Failed to deserialize request", err)
	}

	switch request.GetTopic() {
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
	default:
		return Error.New("Unknown request topic", nil)
	}
	return nil
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
		websocketClient.Send(Message.NewAsync("responseMessage", resultPayload).Serialize())
	default:
		server.mutex.RLock()
		connectedClient := server.connectedClients[page]
		server.mutex.RUnlock()
		if connectedClient == nil {
			return Error.New("Client not found", nil)
		}
		commands := DashboardHelpers.GetCommands(connectedClient.client)
		if commands == nil {
			return Error.New("Client has no commands", nil)
		}
		if !commands[command.Command] {
			return Error.New("Command not found", nil)
		}
		resultPayload, err := connectedClient.executeCommand(command.Command, command.Args)
		if err != nil {
			return Error.New("Failed to execute command", err)
		}
		websocketClient.Send(Message.NewAsync("responseMessage", resultPayload).Serialize())
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
		if !DashboardHelpers.HasStart(connectedClient.client) {
			return Error.New("Client has no start function", nil)
		}
		newStatus, err := connectedClient.executeStart()
		if err != nil {
			return Error.New("Failed to start client", err)
		}
		err = DashboardHelpers.SetStatus(connectedClient.client, newStatus)
		if err != nil {
			return Error.New("Failed to update status", err)
		}
		server.websocketServer.Broadcast(Message.NewAsync("statusUpdate",
			DashboardHelpers.NewStatusUpdate(
				connectedClient.connection.GetName(),
				newStatus,
			).Marshal(),
		))
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
		if !DashboardHelpers.HasStop(connectedClient.client) {
			return Error.New("Client has no stop function", nil)
		}
		newStatus, err := connectedClient.executeStop()
		if err != nil {
			return Error.New("Failed to stop client", err)
		}
		err = DashboardHelpers.SetStatus(connectedClient.client, newStatus)
		if err != nil {
			return Error.New("Failed to update status", err)
		}
		server.websocketServer.Broadcast(Message.NewAsync("statusUpdate",
			DashboardHelpers.NewStatusUpdate(
				connectedClient.connection.GetName(),
				newStatus,
			).Marshal(),
		))
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

func (server *Server) changeWebsocketClientLocation(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
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
	go websocketClient.Send(
		Message.NewAsync("changePage",
			page.Marshal(),
		).Serialize(),
	)
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
