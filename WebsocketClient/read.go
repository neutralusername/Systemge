package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Event"
)

func (connection *WebsocketClient) Read(timeoutMs uint32) ([]byte, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if connection.eventHandler != nil {
		if event := connection.onEvent(Event.New(
			Event.ReadingMessage,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return nil, errors.New("read canceled")
		}
	}

	connection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	_, messageBytes, err := connection.websocketConn.ReadMessage()
	if err != nil {
		if connection.eventHandler != nil {
			connection.onEvent(Event.New(
				Event.ReadMessageFailed,
				Event.Context{
					Event.Error: err.Error(),
				},
				Event.Continue,
			))
		}
		return nil, err
	}
	connection.BytesReceived.Add(uint64(len(messageBytes)))

	if connection.eventHandler != nil {
		if event := connection.onEvent(Event.New(
			Event.ReadMessage,
			Event.Context{
				Event.Bytes: string(messageBytes),
			},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return nil, errors.New("read canceled")
		}
	}
	connection.MessagesReceived.Add(1)

	return messageBytes, nil
}
