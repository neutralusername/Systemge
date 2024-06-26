package Node

import (
	"github.com/gorilla/websocket"
)

func (client *Node) handleWebsocketConnections() {
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

func (client *Node) handleWebsocketConn(websocketConn *websocket.Conn) {
	websocketClient := client.addWebsocketConn(websocketConn)
	client.websocketApplication.OnConnectHandler(client, websocketClient)
	client.handleMessages(websocketClient)
}
