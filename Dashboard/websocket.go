package Dashboard

import (
	"runtime"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Node"
)

func (app *App) GetWebsocketMessageHandlers() map[string]Node.WebsocketMessageHandler {
	return map[string]Node.WebsocketMessageHandler{
		"start": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			app.mutex.Lock()
			n := app.nodes[message.GetPayload()]
			app.mutex.Unlock()
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
		"stop": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			app.mutex.Lock()
			n := app.nodes[message.GetPayload()]
			app.mutex.Unlock()
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
		"reset": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			app.mutex.Lock()
			n := app.nodes[message.GetPayload()]
			app.mutex.Unlock()
			if n == nil {
				return Error.New("Node not found", nil)
			}
			err := n.Reset()
			if err != nil {
				return Error.New("Failed to reset node \""+n.GetName()+"\": "+err.Error(), nil)
			}
			websocketClient.Send(Message.NewAsync("nodeStatus", Helpers.JsonMarshal(NodeStatus{Name: message.GetPayload(), Status: n.GetStatus()})).Serialize())
			return nil
		},
		"command": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
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
		"gc": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			runtime.GC()
			return nil
		},
		"close": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			node.Stop()
			return nil
		},
	}
}

func (app *App) OnConnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	for _, n := range app.nodes {
		go func() {
			websocketClient.Send(Message.NewAsync("nodeStatus", Helpers.JsonMarshal(newNodeStatus(n))).Serialize())
			websocketClient.Send(Message.NewAsync("nodeCommands", Helpers.JsonMarshal(newNodeCommands(n))).Serialize())
		}()
	}
}

func (app *App) OnDisconnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {

}

func (app *App) nodeCommand(command *Command) (string, error) {
	app.mutex.Lock()
	n := app.nodes[command.Name]
	app.mutex.Unlock()
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
