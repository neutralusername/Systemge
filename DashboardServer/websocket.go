package DashboardServer

import (
	"runtime"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

/* func (server *Server) startHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	server.mutex.RLock()
	connectedClient := server.connectedClients[message.GetPayload()]
	server.mutex.RUnlock()
	if connectedClient == nil {
		return Error.New("Client not found", nil)
	}
	if !DashboardHelpers.HasStart(connectedClient.client) {
		return Error.New("Client has no start function", nil)
	}
	response, err := connectedClient.connection.SyncRequestBlocking(Message.TOPIC_START, "")
	if err != nil {
		return Error.New("Failed to send start request to client \""+connectedClient.connection.GetName()+"\": "+err.Error(), nil)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return Error.New(response.GetPayload(), nil)
	}
	err = DashboardHelpers.UpdateStatus(connectedClient.client, Helpers.StringToInt(response.GetPayload()))
	if err != nil {
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Broadcast(Message.NewAsync("statusUpdate",
		Helpers.JsonMarshal(DashboardHelpers.StatusUpdate{
			Name:   connectedClient.connection.GetName(),
			Status: Helpers.StringToInt(response.GetPayload()),
		}),
	))
	return nil
}

func (server *Server) stopHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	server.mutex.RLock()
	connectedClient := server.connectedClients[message.GetPayload()]
	server.mutex.RUnlock()
	if connectedClient == nil {
		return Error.New("Client not found", nil)
	}
	if !DashboardHelpers.HasStop(connectedClient.client) {
		return Error.New("Client has no start function", nil)
	}
	response, err := connectedClient.connection.SyncRequestBlocking(Message.TOPIC_STOP, "")
	if err != nil {
		return Error.New("Failed to send stop request to client \""+connectedClient.connection.GetName()+"\": "+err.Error(), nil)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return Error.New(response.GetPayload(), nil)
	}
	err = DashboardHelpers.UpdateStatus(connectedClient.client, Helpers.StringToInt(response.GetPayload()))
	if err != nil {
		return Error.New("Failed to update status", err)
	}
	server.websocketServer.Broadcast(Message.NewAsync("statusUpdate",
		Helpers.JsonMarshal(DashboardHelpers.StatusUpdate{
			Name:   connectedClient.connection.GetName(),
			Status: Helpers.StringToInt(response.GetPayload()),
		}),
	))
	return nil
}
*/

func (server *Server) gcHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	runtime.GC()
	return nil
}

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
