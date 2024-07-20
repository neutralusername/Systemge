package Dashboard

import (
	"Systemge/Config"
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
		ClientWatchdogTimeoutMs:          20000,
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
				node.Start()
			}
			return nil
		},
	}
}

func (app *App) OnConnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {

}

func (app *App) OnDisconnectHandler(node *Node.Node, websocketClient *Node.WebsocketClient) {

}
