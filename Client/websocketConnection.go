package Client

import (
	"Systemge/Utilities"

	"github.com/gorilla/websocket"
)

func (client *Client) queueNewWebsocketConn(websocketConn *websocket.Conn) error {
	if !client.IsStarted() {
		return Utilities.NewError("websocket listener is not started", nil)
	}
	client.websocketConnChannel <- websocketConn
	return nil
}
