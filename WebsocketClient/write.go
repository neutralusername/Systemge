package WebsocketClient

import (
	"errors"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Event"
)

func (connection *WebsocketClient) Write(messageBytes []byte) error {
	connection.sendMutex.Lock()
	defer connection.sendMutex.Unlock()

	if connection.eventHandler != nil {
		if event := connection.onEvent(Event.New(
			Event.WritingMessage,
			Event.Context{
				Event.Bytes: string(messageBytes),
			},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return errors.New("write canceled")
		}
	}

	err := connection.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		if connection.eventHandler != nil {
			connection.onEvent(Event.New(
				Event.WriteMessageFailed,
				Event.Context{
					Event.Bytes: string(messageBytes),
					Event.Error: err.Error(),
				},
				Event.Cancel,
			))
		}
		return err
	}
	connection.bytesSent.Add(uint64(len(messageBytes)))
	connection.messagesSent.Add(1)

	if connection.eventHandler != nil {
		connection.onEvent(Event.New(
			Event.WroteMessage,
			Event.Context{
				Event.Bytes: string(messageBytes),
			},
			Event.Continue,
		))
	}

	return nil
}
