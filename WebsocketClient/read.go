package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Event"
)

func (connection *WebsocketClient) Read(deadlineMs uint32) ([]byte, error) {
	if connection.eventHandler != nil {
		if event := connection.onEvent(Event.New(
			Event.ReadingMessage,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return nil, errors.New("connection closed")
		}
	}

	connection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(deadlineMs) * time.Millisecond))
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
		connection.Close()
		return nil, err
	}
	connection.bytesReceived.Add(uint64(len(messageBytes)))

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
	connection.messagesReceived.Add(1)

	return messageBytes, nil
}
