package WebsocketClient

import (
	"time"

	"github.com/gorilla/websocket"
)

func (client *WebsocketClient) Write(messageBytes []byte) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	err := client.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		if isWebsocketConnClosedErr(err) {
			client.Close()
		}
	}
	client.BytesSent.Add(uint64(len(messageBytes)))
	client.MessagesSent.Add(1)

	return nil
}

func (client *WebsocketClient) SetWriteDeadline(timeoutMs uint64) error {
	client.websocketConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	return nil
}
