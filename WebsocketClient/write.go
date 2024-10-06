package WebsocketClient

import (
	"time"

	"github.com/gorilla/websocket"
)

func (connection *WebsocketClient) Write(messageBytes []byte, timeoutMs uint32) error {
	connection.sendMutex.Lock()
	defer connection.sendMutex.Unlock()

	connection.websocketConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	err := connection.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		return err
	}
	connection.BytesSent.Add(uint64(len(messageBytes)))
	connection.MessagesSent.Add(1)

	return nil
}
