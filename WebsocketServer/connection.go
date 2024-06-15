package WebsocketServer

import (
	"Systemge/Utilities"

	"github.com/gorilla/websocket"
)

func (server *Server) queueNewWebsocketConn(websocketConn *websocket.Conn) error {
	if !server.IsStarted() {
		return Utilities.NewError("websocket listener is not started", nil)
	}
	server.websocketConnChannel <- websocketConn
	return nil
}
