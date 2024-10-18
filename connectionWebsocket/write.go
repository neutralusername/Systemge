package connectionWebsocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/tools"
)

func (client *WebsocketConnection) WriteChannel(data []byte) <-chan error {
	return tools.ChannelCall(func() (error, error) {
		err := client.Write(data, 0)
		return err, nil
	})
}

func (connection *WebsocketConnection) Write(data []byte, timeoutNs int64) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()

	connection.SetWriteDeadline(timeoutNs)
	err := connection.websocketConn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		if helpers.IsWebsocketConnClosedErr(err) {
			connection.Close()
		}
		return err
	}
	connection.BytesSent.Add(uint64(len(data)))
	connection.MessagesSent.Add(1)
	return nil
}

func (connection *WebsocketConnection) SetWriteDeadline(timeoutNs int64) {
	connection.websocketConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
}
