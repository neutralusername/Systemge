package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
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
	listener.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	listener.status = Status.Pending
	listener.stopChannel = make(chan struct{})

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
		close(listener.stopChannel)
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
