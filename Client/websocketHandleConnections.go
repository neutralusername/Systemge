package Client

import (
	"github.com/gorilla/websocket"
)

func (client *Client) handleWebsocketConnections() {
	for client.IsStarted() {
		select {
		case <-client.stopChannel:
			return
		case websocketConn := <-client.websocketConnChannel:
			if websocketConn == nil {
				continue
			}
			go client.handleWebsocketConn(websocketConn)
		}
	}
}

func (client *Client) handleWebsocketConn(websocketConn *websocket.Conn) {
	websocketClient := client.addWebsocketConn(websocketConn)
	client.websocketApplication.OnConnectHandler(client, websocketClient)
	client.handleMessages(websocketClient)
}
