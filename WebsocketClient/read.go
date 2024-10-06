package WebsocketClient

import (
	"time"
)

func (client *WebsocketClient) Read(timeoutMs uint32) ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	client.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	_, messageBytes, err := client.websocketConn.ReadMessage()
	if err != nil {
		return nil, err
	}
	client.BytesReceived.Add(uint64(len(messageBytes)))
	client.MessagesReceived.Add(1)

	return messageBytes, nil
}
