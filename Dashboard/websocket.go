package Dashboard

import (
	"net/http"
	"runtime"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Node"

	"github.com/gorilla/websocket"
)

func (app *App) GetWebsocketComponentConfig() *Config.Websocket {
	return &Config.Websocket{
		Pattern: "/ws",
		ServerConfig: &Config.TcpServer{
			Port:        18251,
			TlsCertPath: app.config.ServerConfig.TlsCertPath,
			TlsKeyPath:  app.config.ServerConfig.TlsKeyPath,
			Blacklist:   app.config.ServerConfig.Blacklist,
			Whitelist:   app.config.ServerConfig.Whitelist,
		},
		HandleClientMessagesSequentially: true,
		ClientMessageCooldownMs:          0,
		ClientWatchdogTimeoutMs:          1000 * 60 * 5,
		Upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (app *App) GetWebsocketMessageHandlers() map[string]Node.WebsocketMessageHandler {
	return map[string]Node.WebsocketMessageHandler{
		"start": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			err := app.startNode(message.GetPayload())
			if err != nil {
				return err
			}
			websocketClient.Send(Message.NewAsync("nodeStatus", Helpers.JsonMarshal(NodeStatus{Name: message.GetPayload(), Status: true})).Serialize())
			return nil
		},
		"stop": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			err := app.stopNode(message.GetPayload())
			if err != nil {
				return err
			}
			websocketClient.Send(Message.NewAsync("nodeStatus", Helpers.JsonMarshal(NodeStatus{Name: message.GetPayload(), Status: false})).Serialize())
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

func (app *App) startNode(nodeName string) error {
	app.mutex.Lock()
	n := app.nodes[nodeName]
	app.mutex.Unlock()
	if n == nil {
		return Error.New("Node not found", nil)
	}
	err := n.Start()
	if err != nil {
		return Error.New("Failed to start node \""+n.GetName()+"\": "+err.Error(), nil)
	}
	return nil
}

func (app *App) stopNode(nodeName string) error {
	app.mutex.Lock()
	n := app.nodes[nodeName]
	app.mutex.Unlock()
	if n == nil {
		return Error.New("Node not found", nil)
	}
	err := n.Stop()
	if err != nil {
		return Error.New("Failed to stop node \""+n.GetName()+"\": "+err.Error(), nil)
	}
	return nil
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
