package WebsocketServer

import (
	"github.com/gorilla/websocket"
)

func (server *Server) handleWebsocketConn(websocketConn *websocket.Conn) {
	client := server.addWebsocketConn(websocketConn)
	server.websocketApplication.OnConnectHandler(client)
	server.handleMessages(client)
}
