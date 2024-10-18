package connectionWebsocket

import (
	"time"

	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/tools"
)

func (client *WebsocketConnection) ReadChannel() <-chan []byte {
	return tools.ChannelCall(func() ([]byte, error) {
		return client.Read(0)
	})
}

func (connection *WebsocketConnection) Read(timeoutNs int64) ([]byte, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	connection.SetReadDeadline(timeoutNs)
	_, data, err := connection.websocketConn.ReadMessage()
	if err != nil {
		if helpers.IsWebsocketConnClosedErr(err) {
			connection.Close()
		}
		return nil, err
	}
	connection.BytesReceived.Add(uint64(len(data)))
	connection.MessagesReceived.Add(1)
	return data, nil
}

func (connection *WebsocketConnection) SetReadDeadline(timeoutMs int64) {
	connection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}

func (connection *WebsocketConnection) SetReadLimit(maxBytes int64) {
	connection.websocketConn.SetReadLimit(maxBytes)
}
