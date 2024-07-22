package Dashboard

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Helpers"
	"Systemge/Message"
	"Systemge/Node"
	"net/http"
	"runtime"

	"github.com/gorilla/websocket"
)

func (app *App) GetWebsocketComponentConfig() *Config.Websocket {
	return &Config.Websocket{
		Pattern: "/ws",
		Server: &Config.TcpServer{
			Port:        18251,
			TlsCertPath: app.config.Server.TlsCertPath,
			TlsKeyPath:  app.config.Server.TlsKeyPath,
			Blacklist:   app.config.Server.Blacklist,
			Whitelist:   app.config.Server.Whitelist,
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
			n := app.nodes[message.GetPayload()]
			err := n.Start()
			if err != nil {
				return Error.New("Failed to start node \""+n.GetName()+"\": "+err.Error(), nil)
			}
			websocketClient.Send(Message.NewAsync("nodeStatus", node.GetName(), Helpers.JsonMarshal(newNodeStatus(n))).Serialize())
			return nil
		},
		"stop": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			n := app.nodes[message.GetPayload()]
			err := n.Stop()
			if err != nil {
				return Error.New("Failed to stop node \""+n.GetName()+"\": "+err.Error(), nil)
			}
			websocketClient.Send(Message.NewAsync("nodeStatus", node.GetName(), Helpers.JsonMarshal(newNodeStatus(n))).Serialize())
			return nil
		},
		"command": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			command := jsonUnmarshal(message.GetPayload())
			if command == nil {
				return Error.New("Invalid command", nil)
			}
			n := app.nodes[command.Name]
			if n == nil {
				return Error.New("Node not found", nil)
			}
			commandHandler := n.GetCommandHandlers()[command.Command]
			if commandHandler == nil {
				return Error.New("Command not found", nil)
			}
			result, err := commandHandler(n, command.Args)
			if err != nil {
				return Error.New("Failed to execute command \""+command.Command+"\": "+err.Error(), nil)
			}
			websocketClient.Send(Message.NewAsync("responseMessage", node.GetName(), result).Serialize())
			return nil
		},
		"gc": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			runtime.GC()
			return nil
		},
	}
}

func (app *App) OnConnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {
	for _, n := range app.nodes {
		websocketClient.Send(Message.NewAsync("nodeStatus", node.GetName(), Helpers.JsonMarshal(newNodeStatus(n))).Serialize())
		websocketClient.Send(Message.NewAsync("nodeCommands", node.GetName(), Helpers.JsonMarshal(newNodeCommands(n))).Serialize())
	}
}

func (app *App) OnDisconnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {

}
