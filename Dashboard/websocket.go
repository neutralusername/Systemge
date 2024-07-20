package Dashboard

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Node"
	"net/http"

	"github.com/gorilla/websocket"
)

func (app *App) GetWebsocketComponentConfig() *Config.Websocket {
	return &Config.Websocket{
		Pattern: "/ws",
		Server: &Config.TcpServer{
			Port: 18251,
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
				app.node.GetErrorLogger().Log(Error.New("Failed to start node \""+n.GetName()+"\": "+err.Error(), nil).Error())
			} else {
				websocketClient.Send(Message.NewAsync("nodeStatus", node.GetName(), jsonMarshal(newNodeStatus(n.GetName(), n.IsStarted()))).Serialize())
			}
			return nil
		},
		"stop": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			n := app.nodes[message.GetPayload()]
			err := n.Stop()
			if err != nil {
				app.node.GetErrorLogger().Log(Error.New("Failed to stop node \""+n.GetName()+"\": "+err.Error(), nil).Error())
			} else {
				websocketClient.Send(Message.NewAsync("nodeStatus", node.GetName(), jsonMarshal(newNodeStatus(n.GetName(), n.IsStarted()))).Serialize())
			}
			return nil
		},
	}
}

func (app *App) OnConnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {
	for _, n := range app.nodes {
		websocketClient.Send(Message.NewAsync("nodeStatus", node.GetName(), jsonMarshal(newNodeStatus(n.GetName(), n.IsStarted()))).Serialize())
	}
}

func (app *App) OnDisconnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {

}
