package Node

import (
	"Systemge/Error"

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
	node.GetLogger().Info(Error.New("websocket client connected with id \""+websocketClient.GetId()+"\" and ip \""+websocketClient.GetIp()+"\" on node \""+node.GetName()+"\"", nil).Error())
	node.GetWebsocketComponent().OnConnectHandler(node, websocketClient)
	node.handleMessages(websocketClient)
}
