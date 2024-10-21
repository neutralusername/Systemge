package connectionWebsocket

import (
	"time"

	"github.com/neutralusername/systemge/helpers"
)

func (connection *WebsocketConnection) Read(timeoutNs int64) ([]byte, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	connection.SetReadDeadline(timeoutNs)
	_, data, err := connection.websocketConn.ReadMessage()
	if err != nil {

		if helpers.IsWebsocketConnClosedErr(err) {
			println("test")
			connection.Close()
		}
		return nil, err
	}
	connection.BytesReceived.Add(uint64(len(data)))
	connection.MessagesReceived.Add(1)
	return data, nil
}

func (connection *WebsocketConnection) SetReadDeadline(timeoutNs int64) {
	if timeoutNs == 0 {
		connection.websocketConn.SetReadDeadline(time.Time{})
		return
	}
	connection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
}

func (connection *WebsocketConnection) SetReadLimit(maxBytes int64) {
	connection.websocketConn.SetReadLimit(maxBytes)
}
