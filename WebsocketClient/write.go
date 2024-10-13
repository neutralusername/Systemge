package WebsocketClient

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Helpers"
)

func (client *WebsocketClient) write(messageBytes []byte) error {
	err := client.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		if Helpers.IsWebsocketConnClosedErr(err) {
			client.Close()
		}
		return err
	}
	client.BytesSent.Add(uint64(len(messageBytes)))
	client.MessagesSent.Add(1)
	return nil
}

func (client *WebsocketClient) Write(messageBytes []byte) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	return client.write(messageBytes)
}

func (client *WebsocketClient) WriteTimeout(messageBytes []byte, timeoutMs uint64) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	client.SetWriteDeadline(timeoutMs)
	return client.write(messageBytes)
}

func (client *WebsocketClient) SetWriteDeadline(timeoutMs uint64) {
	client.websocketConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}
