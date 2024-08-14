package Dashboard

import (
	"runtime"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (app *App) GetWebsocketMessageHandlers() map[string]WebsocketServer.WebsocketMessageHandler {
	return map[string]WebsocketServer.WebsocketMessageHandler{
		"start": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			app.mutex.RLock()
			n := app.nodes[message.GetPayload()]
			app.mutex.RUnlock()
			if n == nil {
				return Error.New("Node not found", nil)
			}
			err := n.Start()
			if err != nil {
				return Error.New("Failed to start node \""+n.GetName()+"\": "+err.Error(), nil)
			}
			websocketClient.Send(Message.NewAsync("nodeStatus", Helpers.JsonMarshal(NodeStatus{Name: message.GetPayload(), Status: n.GetStatus()})).Serialize())
			return nil
		},
		"stop": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			app.mutex.RLock()
			n := app.nodes[message.GetPayload()]
			app.mutex.RUnlock()
			if n == nil {
				return Error.New("Node not found", nil)
			}
			err := n.Stop()
			if err != nil {
				return Error.New("Failed to stop node \""+n.GetName()+"\": "+err.Error(), nil)
			}
			websocketClient.Send(Message.NewAsync("nodeStatus", Helpers.JsonMarshal(NodeStatus{Name: message.GetPayload(), Status: n.GetStatus()})).Serialize())
			return nil
		},
		"command": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			command := unmarshalCommand(message.GetPayload())
			if command == nil {
				return Error.New("Invalid command", nil)
			}
			result, err := app.nodeCommand(command)
			if err != nil {
				return err
			}
			websocketClient.Send(Message.NewAsync("responseMessage", result).Serialize())
			return nil
		},
		"gc": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			runtime.GC()
			return nil
		},
	}
}

func (app *App) OnConnectHandler(websocketClient *WebsocketServer.WebsocketClient) {
	app.mutex.RLock()
	defer app.mutex.RUnlock()
	for _, n := range app.nodes {
		go func() {
			websocketClient.Send(Message.NewAsync("addNode", Helpers.JsonMarshal(newAddNode(n))).Serialize())
		}()
	}
}

func (app *App) OnDisconnectHandler(websocketClient *WebsocketServer.WebsocketClient) {

}

func (app *App) nodeCommand(command *Command) (string, error) {
	app.mutex.RLock()
	n := app.nodes[command.Name]
	app.mutex.RUnlock()
	if n == nil {
		return "", Error.New("Node not found", nil)
	}
	commandHandler := n.GetCommandHandlers()[command.Command]
	if commandHandler == nil {
		return "", Error.New("Command not found", nil)
	}
	result, err := commandHandler(n, command.Args)
	if err != nil {
		return "", Error.New("Failed to execute command \""+command.Name+"\": "+err.Error(), nil)
	}
	return result, nil
}
