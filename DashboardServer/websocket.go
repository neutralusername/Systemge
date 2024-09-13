package DashboardServer

import (
	"runtime"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (server *Server) gcHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	runtime.GC()
	return nil
}

func (server *Server) startHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
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

func (server *Server) commandHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
	if err != nil {
		return err
	}
	resultPayload := ""
	switch command.Name {
	case "":
		resultPayload, err = server.dashboardCommandHandler(command)
		if err != nil {
			return Error.New("Failed to execute command", err)
		}
	default:
		server.mutex.RLock()
		client := server.connectedClients[command.Name]
		server.mutex.RUnlock()
		if client == nil {
			return Error.New("Client not found", nil)
		}
		resultPayload, err = client.ExecuteCommand(command.Command, command.Args)
		if err != nil {
			return Error.New("Failed to execute command", err)
		}
	}
	websocketClient.Send(Message.NewAsync("responseMessage", resultPayload).Serialize())
	return nil
}
func (server *Server) dashboardCommandHandler(command *DashboardHelpers.Command) (string, error) {
	commandHandler, _ := server.dashboardCommandHandlers.Get(command.Command)
	if commandHandler == nil {
		return "", Error.New("Command not found", nil)
	}
	return commandHandler(command.Args)
}

func (server *Server) changeWebsocketClientLocation(client *WebsocketServer.WebsocketClient, message *Message.Message) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	switch message.GetPayload() {
	case "":
		go client.Send(
			Message.NewAsync("changePage",
				DashboardHelpers.NewPageUpdate(
					server.getDashboardData(),
					DashboardHelpers.PAGE_DASHBOARD,
				).Marshal(),
			).Serialize(),
		)
	default:
		connectedClient := server.connectedClients[message.GetPayload()]
		if connectedClient == nil {
			return Error.New("Client not found", nil)
		}
		switch connectedClient.client.(type) {
		case *DashboardHelpers.CustomServiceClient:
			go client.Send(
				Message.NewAsync("changePage",
					DashboardHelpers.NewPageUpdate(
						connectedClient.client,
						DashboardHelpers.PAGE_CUSTOMSERVICE,
					).Marshal(),
				).Serialize(),
			)
		case *DashboardHelpers.CommandClient:
			go client.Send(
				Message.NewAsync("changePage",
					DashboardHelpers.NewPageUpdate(
						connectedClient.client,
						DashboardHelpers.PAGE_COMMAND,
					).Marshal(),
				).Serialize(),
			)
		}
	}
	server.websocketClientLocations[client.GetId()] = message.GetPayload()
	return nil
}
func (server *Server) getDashboardData() map[string]interface{} {
	dashboardData := map[string]interface{}{}
	clientsSlice := []interface{}{}
	dashboardData["clients"] = clientsSlice
	for _, connectedClient := range server.connectedClients {
		dashboardData["clients"] = append(clientsSlice, connectedClient.client)
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
	server.websocketClientLocations[websocketClient.GetId()] = ""
	return nil
}
func (server *Server) onWebsocketDisconnectHandler(websocketClient *WebsocketServer.WebsocketClient) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.websocketClientLocations, websocketClient.GetId())
}
