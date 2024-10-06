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
			Event.Context{
				Event.Circumstance: Event.ServiceStoping,
			},
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
				Event.Context{
					Event.Circumstance: Event.ServiceStoping,
				},
				Event.Continue,
			))
		}
		return errors.New("websocketListener is already stopped")
	}

	listener.status = Status.Pending
	listener.httpServer.Stop()
	if listener.ipRateLimiter != nil {
		listener.ipRateLimiter.Close()
		listener.ipRateLimiter = nil
	}

	if listener.eventHandler != nil {
		listener.onEvent(Event.New(
			Event.ServiceStoped,
			Event.Context{
				Event.Circumstance: Event.ServiceStoping,
			},
			Event.Continue,
		))
	}

	listener.status = Status.Stopped
	return nil
}
