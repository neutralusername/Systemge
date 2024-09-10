package DashboardServer

import (
	"runtime"

	"github.com/neutralusername/Systemge/DashboardUtilities"
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
	app.websocketServer.Broadcast(Message.NewAsync("statusUpdate", Helpers.JsonMarshal(statusUpdate{Name: client.Name, Status: client.Status})))
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
	app.websocketServer.Broadcast(Message.NewAsync("statusUpdate", Helpers.JsonMarshal(statusUpdate{Name: client.Name, Status: client.Status})))
	return nil
}

func (app *Server) gcHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	runtime.GC()
	return nil
}

func (app *Server) commandHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	command, err := DashboardUtilities.UnmarshalCommand(message.GetPayload())
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
	app.mutex.RLock()
	defer app.mutex.RUnlock()
	for _, client := range app.clients {
		go func() {
			websocketClient.Send(Message.NewAsync("addModule", Helpers.JsonMarshal(client)).Serialize())
		}()
	}
	commands := []string{}
	for command := range app.commandHandlers {
		commands = append(commands, command)
	}
	websocketClient.Send(Message.NewAsync("dashboardCommands", Helpers.JsonMarshal(commands)).Serialize())
	return nil
}
