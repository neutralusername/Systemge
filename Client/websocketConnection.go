package Client

import (
	"Systemge/Error"

	"github.com/gorilla/websocket"
)

func (client *Client) queueNewWebsocketConn(websocketConn *websocket.Conn) error {
	if !client.IsStarted() {
		return Error.New("websocket listener is not started", nil)
	}
	client.websocketConnChannel <- websocketConn
	return nil
}
