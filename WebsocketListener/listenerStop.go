package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (listener *WebsocketListener) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.eventHandler != nil {
		if event := listener.onEvent(Event.New(
			Event.ServiceStoping,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return errors.New("stop canceled")
		}
	}

	if listener.status == Status.Stopped {
		if listener.eventHandler != nil {
			listener.onEvent(Event.New(
				Event.ServiceAlreadyStoped,
				Event.Context{},
				Event.Continue,
			))
		}
		return errors.New("websocketListener is already stopped")
	}

	listener.status = Status.Pending
	close(listener.stopChannel)

	if err := listener.httpServer.Stop(); err != nil {
		if listener.eventHandler != nil {
			listener.onEvent(Event.New(
				Event.ServiceStopFailed,
				Event.Context{
					Event.Error: err.Error(),
				},
				Event.Continue,
			))
		}
		listener.status = Status.Started
	}

	if listener.eventHandler != nil {
		listener.onEvent(Event.New(
			Event.ServiceStoped,
			Event.Context{},
			Event.Continue,
		))
	}

	listener.status = Status.Stopped
	return nil
}
