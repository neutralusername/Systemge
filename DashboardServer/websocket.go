package DashboardServer

import (
	"runtime"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (app *Server) startHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	app.mutex.RLock()
	client := app.clients[message.GetPayload()]
	app.mutex.RUnlock()
	if client == nil {
		return Error.New("Client not found", nil)
	}
	if !client.HasStartFunc {
		return Error.New("Client has no start function", nil)
	}
	response, err := client.Connection.SyncRequestBlocking(Message.TOPIC_START, "")
	if err != nil {
		return Error.New("Failed to send start request to client \""+client.Name+"\": "+err.Error(), nil)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return Error.New(response.GetPayload(), nil)
	}
	client.Status = Helpers.StringToInt(response.GetPayload())
	app.websocketServer.Broadcast(Message.NewAsync("statusUpdate", Helpers.JsonMarshal(DashboardHelpers.StatusUpdate{Name: client.Name, Status: client.Status})))
	return nil
}

func (app *Server) stopHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	app.mutex.RLock()
	client := app.clients[message.GetPayload()]
	app.mutex.RUnlock()
	if client == nil {
		return Error.New("Client not found", nil)
	}
	if !client.HasStopFunc {
		return Error.New("Client has no stop function", nil)
	}
	response, err := client.Connection.SyncRequestBlocking(Message.TOPIC_STOP, "")
	if err != nil {
		return Error.New("Failed to send stop request to client \""+client.Name+"\": "+err.Error(), nil)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return Error.New(response.GetPayload(), nil)
	}
	client.Status = Helpers.StringToInt(response.GetPayload())
	app.websocketServer.Broadcast(Message.NewAsync("statusUpdate", Helpers.JsonMarshal(DashboardHelpers.StatusUpdate{Name: client.Name, Status: client.Status})))
	return nil
}

func (app *Server) gcHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	runtime.GC()
	return nil
}

func (app *Server) commandHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	command, err := DashboardHelpers.UnmarshalCommand(message.GetPayload())
	if err != nil {
		return err
	}

	switch command.Name {
	case "dashboard":
		result, err := app.dashboardCommandHandler(command)
		if err != nil {
			return Error.New("Failed to execute command", err)
		}
		websocketClient.Send(Message.NewAsync("responseMessage", result).Serialize())
	default:
		app.mutex.RLock()
		client := app.clients[command.Name]
		app.mutex.RUnlock()
		if client == nil {
			return Error.New("Client not found", nil)
		}
		result, err := client.ExecuteCommand(command.Command, command.Args)
		if err != nil {
			return Error.New("Failed to execute command", err)
		}
		websocketClient.Send(Message.NewAsync("responseMessage", result).Serialize())
	}
	return nil
}

func (app *Server) onWebsocketConnectHandler(websocketClient *WebsocketServer.WebsocketClient) error {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	app.websocketClientLocations[websocketClient.GetId()] = ""

	/* for _, client := range app.clients {
		go func() {
			websocketClient.Send(Message.NewAsync("addModule", Helpers.JsonMarshal(client)).Serialize())
		}()
	}
	commands := []string{}
	for command := range app.commandHandlers {
		commands = append(commands, command)
	}
	websocketClient.Send(Message.NewAsync("dashboardCommands", Helpers.JsonMarshal(commands)).Serialize()) */
	return nil
}

func (app *Server) onWebsocketDisconnectHandler(websocketClient *WebsocketServer.WebsocketClient) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	delete(app.websocketClientLocations, websocketClient.GetId())
}

func (app *Server) dashboardCommandHandler(command *DashboardHelpers.Command) (string, error) {
	commandHandler, _ := app.commandHandlers.Get(command.Command)
	if commandHandler == nil {
		return "", Error.New("Command not found", nil)
	}
	return commandHandler(command.Args)
}
