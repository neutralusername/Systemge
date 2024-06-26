package Node

import (
	"github.com/gorilla/websocket"
)

func (node *Node) handleWebsocketConnections() {
	for node.IsStarted() {
		select {
		case <-node.stopChannel:
			return
		case websocketConn := <-node.websocketConnChannel:
			if websocketConn == nil {
				continue
			}
			go node.handleWebsocketConn(websocketConn)
		}
	}
}

func (node *Node) handleWebsocketConn(websocketConn *websocket.Conn) {
	websocketClient := node.addWebsocketConn(websocketConn)
	node.websocketComponent.OnConnectHandler(node, websocketClient)
	node.handleMessages(websocketClient)
}
