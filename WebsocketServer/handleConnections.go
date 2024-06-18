package WebsocketServer

import (
	"github.com/gorilla/websocket"
)

func (server *Server) handleWebsocketConnections() {
	for server.IsStarted() {
		select {
		case <-server.stopChannel:
			return
		case websocketConn := <-server.websocketConnChannel:
			if websocketConn == nil {
				continue
			}
			go server.handleWebsocketConn(websocketConn)
		}
	}
}

func (server *Server) handleWebsocketConn(websocketConn *websocket.Conn) {
	client := server.addWebsocketConn(websocketConn)
	server.websocketApplication.OnConnectHandler(client)
	server.handleMessages(client)
}
