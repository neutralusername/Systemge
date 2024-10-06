package WebsocketClient

import (
	"time"
)

func (connection *WebsocketClient) Read(timeoutMs uint32) ([]byte, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	connection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	_, messageBytes, err := connection.websocketConn.ReadMessage()
	if err != nil {
		return nil, err
	}
	connection.BytesReceived.Add(uint64(len(messageBytes)))
	connection.MessagesReceived.Add(1)

	return messageBytes, nil
}
