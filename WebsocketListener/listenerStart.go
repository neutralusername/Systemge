package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (listener *WebsocketListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.eventHandler != nil {
		if event := listener.onEvent(Event.New(
			Event.ServiceStarting,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return errors.New("start canceled")
		}
	}

	if listener.status == Status.Started {
		if listener.eventHandler != nil {
			listener.onEvent(Event.New(
				Event.ServiceAlreadyStarted,
				Event.Context{},
				Event.Cancel,
			))
		}
		return errors.New("tcpSystemgeListener is already started")
	}

	listener.status = Status.Pending

	if err := listener.httpServer.Start(); err != nil {
		if listener.eventHandler != nil {
			listener.onEvent(Event.New(
				Event.ServiceStartFailed,
				Event.Context{
					Event.Error: err.Error(),
				},
				Event.Cancel,
			))
		}
		listener.status = Status.Stopped
		return err
	}

	listener.status = Status.Started

	if listener.eventHandler != nil {
		listener.onEvent(Event.New(
			Event.ServiceStarted,
			Event.Context{},
			Event.Continue,
		))
	}
	return nil
}
