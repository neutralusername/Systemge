package Dashboard

import (
	"runtime"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (app *DashboardServer) GetWebsocketMessageHandlers() map[string]WebsocketServer.MessageHandler {
	return map[string]WebsocketServer.MessageHandler{
		"start": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			app.mutex.RLock()
			client := app.clients[message.GetPayload()]
			app.mutex.RUnlock()
			if client == nil {
				return Error.New("Node not found", nil)
			}
			if !client.HasStartFunc {
				return Error.New("Node has no start function", nil)
			}
			response, err := client.connection.SyncRequest(Message.TOPIC_START, "")
			if err != nil {
				return Error.New("Failed to send start request to client \""+client.Name+"\": "+err.Error(), nil)
			}
			if response.GetTopic() == Message.TOPIC_FAILURE {
				return Error.New(response.GetPayload(), nil)
			}
			client.Status = Helpers.StringToInt(response.GetPayload())
			app.websocketServer.Broadcast(Message.NewAsync("statusUpdate", Helpers.JsonMarshal(StatusUpdate{Name: client.Name, Status: client.Status})))
			return nil
		},
		"stop": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			app.mutex.RLock()
			client := app.clients[message.GetPayload()]
			app.mutex.RUnlock()
			if client == nil {
				return Error.New("Node not found", nil)
			}
			if !client.HasStopFunc {
				return Error.New("Node has no stop function", nil)
			}
			response, err := client.connection.SyncRequest(Message.TOPIC_STOP, "")
			if err != nil {
				return Error.New("Failed to send stop request to client \""+client.Name+"\": "+err.Error(), nil)
			}
			if response.GetTopic() == Message.TOPIC_FAILURE {
				return Error.New(response.GetPayload(), nil)
			}
			client.Status = Helpers.StringToInt(response.GetPayload())
			app.websocketServer.Broadcast(Message.NewAsync("statusUpdate", Helpers.JsonMarshal(StatusUpdate{Name: client.Name, Status: client.Status})))
			return nil
		},
		"command": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			return nil
		},
		"gc": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			runtime.GC()
			return nil
		},
	}
}

func (app *DashboardServer) OnConnectHandler(websocketClient *WebsocketServer.WebsocketClient) error {
	app.mutex.RLock()
	defer app.mutex.RUnlock()
	for _, client := range app.clients {
		go func() {
			websocketClient.Send(Message.NewAsync("addModule", Helpers.JsonMarshal(client)).Serialize())
		}()
	}
	return nil
}

func (app *DashboardServer) executeCommand(clientName, commandName string, args []string) (string, error) {
	app.mutex.RLock()
	client := app.clients[clientName]
	app.mutex.RUnlock()
	if client == nil {
		return "", Error.New("Node not found", nil)
	}
	response, err := client.executeCommand(commandName, args)
	if err != nil {
		return "", err
	}
	return response.GetPayload(), nil
}
