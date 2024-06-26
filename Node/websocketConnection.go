package Node

import (
	"Systemge/Error"

	"github.com/gorilla/websocket"
)

func (node *Node) queueNewWebsocketConn(websocketConn *websocket.Conn) error {
	if !node.IsStarted() {
		return Error.New("websocket listener is not started", nil)
	}
	node.websocketConnChannel <- websocketConn
	return nil
}
