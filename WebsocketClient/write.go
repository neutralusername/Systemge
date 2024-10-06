package WebsocketClient

import (
	"time"

	"github.com/gorilla/websocket"
)

func (client *WebsocketClient) Write(messageBytes []byte, timeoutMs uint32) error {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()

	client.websocketConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	err := client.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		return err
	}
	client.BytesSent.Add(uint64(len(messageBytes)))
	client.MessagesSent.Add(1)

	return nil
}
