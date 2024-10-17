package connectionWebsocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/systemge/helpers"
)

func (connection *WebsocketConnection) Write(messageBytes []byte, timeoutNs int64) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()

	connection.SetWriteDeadline(timeoutNs)
	err := connection.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		if helpers.IsWebsocketConnClosedErr(err) {
			connection.Close()
		}
		return err
	}
	connection.BytesSent.Add(uint64(len(messageBytes)))
	connection.MessagesSent.Add(1)
	return nil
}

func (connection *WebsocketConnection) SetWriteDeadline(timeoutNs int64) {
	connection.websocketConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
}
