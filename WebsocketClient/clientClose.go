package WebsocketClient

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
)

func (connection *WebsocketClient) Close() error {
	if !connection.closedMutex.TryLock() {
		return errors.New("websocketClient already closing")
	}
	defer connection.closedMutex.Unlock()

	if connection.eventHandler != nil {
		if event := connection.onEvent(Event.New(
			Event.ServiceStoping,
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return errors.New("close canceled")
		}
	}

	if connection.closed {
		if connection.eventHandler != nil {
			connection.onEvent(Event.New(
				Event.ServiceAlreadyStarted,
				Event.Context{
					Event.Circumstance: Event.ServiceStop,
				},
				Event.Cancel,
			))
		}
		return errors.New("websocketClient already closed")
	}

	connection.closed = true
	connection.websocketConn.Close()
	close(connection.closeChannel)

	if connection.eventHandler != nil {
		connection.onEvent(Event.New(
			Event.ServiceStoped,
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
			Event.Continue,
		))
	}

	return nil
}
