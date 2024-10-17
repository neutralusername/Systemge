package connectionWebsocket

import (
	"time"

	"github.com/neutralusername/Systemge/helpers"
)

func (connection *WebsocketConnection) Read(timeoutNs int64) ([]byte, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	connection.SetReadDeadline(timeoutNs)
	_, messageBytes, err := connection.websocketConn.ReadMessage()
	if err != nil {
		if helpers.IsWebsocketConnClosedErr(err) {
			connection.Close()
		}
		return nil, err
	}
	connection.BytesReceived.Add(uint64(len(messageBytes)))
	connection.MessagesReceived.Add(1)
	return messageBytes, nil
}

func (connection *WebsocketConnection) SetReadDeadline(timeoutMs int64) {
	connection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}

func (connection *WebsocketConnection) SetReadLimit(maxBytes int64) {
	connection.websocketConn.SetReadLimit(maxBytes)
}
