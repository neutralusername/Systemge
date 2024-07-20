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
			for _, node := range app.nodes {
				err := node.Start()
				if err != nil {
					app.node.GetErrorLogger().Log(Error.New("Failed to start node \""+node.GetName()+"\": "+err.Error(), nil).Error())
				}
			}
			return nil
		},
		"stop": func(node *Node.Node, websocketClient *Node.WebsocketClient, message *Message.Message) error {
			for _, node := range app.nodes {
				err := node.Stop()
				if err != nil {
					app.node.GetErrorLogger().Log(Error.New("Failed to stop node \""+node.GetName()+"\": "+err.Error(), nil).Error())
				}
			}
			return nil

		},
	}
}

func (app *App) OnConnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {

}

func (app *App) OnDisconnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {

}
