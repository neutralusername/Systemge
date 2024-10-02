package WebsocketClient

import (
	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *WebsocketClient) Write(messageBytes []byte) error {
	return connection.write(messageBytes, Event.WriteRuntime)
}

func (connection *WebsocketClient) write(messageBytes []byte, circumstance string) error {
	connection.sendMutex.Lock()
	defer connection.sendMutex.Unlock()

	if event := connection.onEvent(Event.NewInfo(
		Event.WritingMessage,
		"sending message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: circumstance,
			Event.Bytes:        string(messageBytes),
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	err := connection.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		connection.onEvent(Event.NewWarningNoOption(
			Event.WriteMessageFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: circumstance,
				Event.Bytes:        string(messageBytes),
			}),
		)
		if Tcp.IsConnectionClosed(err) {
			connection.Close()
		}
		return err
	}
	connection.bytesSent.Add(uint64(len(messageBytes)))

	connection.onEvent(Event.NewInfoNoOption(
		Event.WroteMessage,
		"message sent",
		Event.Context{
			Event.Circumstance: circumstance,
			Event.Bytes:        string(messageBytes),
		},
	))

	return nil
}
