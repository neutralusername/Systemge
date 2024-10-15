package ConnectionWebsocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Helpers"
)

func (connection *WebsocketConnection) Write(messageBytes []byte, timeoutMs uint64) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()

	connection.SetWriteDeadline(timeoutMs)
	err := connection.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		if Helpers.IsWebsocketConnClosedErr(err) {
			connection.Close()
		}
		return err
	}
	connection.BytesSent.Add(uint64(len(messageBytes)))
	connection.MessagesSent.Add(1)
	return nil
}

func (connection *WebsocketConnection) SetWriteDeadline(timeoutMs uint64) {
	connection.websocketConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}
