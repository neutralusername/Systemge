package Node

import (
	"Systemge/Error"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// Promotes a http-websocket-handshake request to a websocket connection and queues it for handling on the websocket server
func (client *Node) promoteToWebsocket() func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(httpRequest *http.Request) bool {
			return true
		},
	}
	return func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		websocketConn, err := upgrader.Upgrade(responseWriter, httpRequest, nil)
		if err != nil {
			client.logger.Log(Error.New(fmt.Sprintf("Error upgrading connection to websocket: %s", err.Error()), nil).Error())
			return
		}
		err = client.queueNewWebsocketConn(websocketConn)
		if err != nil {
			client.logger.Log(Error.New("Error queuing new websocket connection", err).Error())
		}
	}
}
