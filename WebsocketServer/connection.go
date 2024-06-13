package WebsocketServer

import (
	"Systemge/Error"

	"github.com/gorilla/websocket"
)

func (server *Server) queueNewWebsocketConn(websocketConn *websocket.Conn) error {
	if !server.IsStarted() {
		return Error.New("websocket listener is not started", nil)
	}
	server.websocketConnChannel <- websocketConn
	return nil
}
