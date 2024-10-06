package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (listener *WebsocketListener) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if event := listener.onEvent(Event.NewInfo(
		Event.ServiceStoping,
		"stopping websocketListener",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStoping,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if listener.status == Status.Stopped {
		listener.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStoped,
			"websocketListener is already stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStoping,
			},
		))
		return errors.New("websocketListener is already stopped")
	}

	listener.status = Status.Pending
	listener.httpServer.Stop()
	if listener.ipRateLimiter != nil {
		listener.ipRateLimiter.Close()
		listener.ipRateLimiter = nil
	}

	listener.onEvent(Event.NewInfoNoOption(
		Event.ServiceStoped,
		"websocketListener stopped",
		Event.Context{
			Event.Circumstance: Event.ServiceStoping,
		},
	))

	listener.status = Status.Stopped
	return nil
}
