package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *WebsocketListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if event := listener.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"starting tcpSystemgeListener",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStarting,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if listener.status == Status.Started {
		listener.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"tcpSystemgeListener is already started",
			Event.Context{
				Event.Circumstance: Event.ServiceStarting,
			},
		))
		return errors.New("tcpSystemgeListener is already started")
	}

	listener.status = Status.Pending
	if listener.config.IpRateLimiter != nil {
		listener.ipRateLimiter = Tools.NewIpRateLimiter(listener.config.IpRateLimiter)
	}
	if err := listener.httpServer.Start(); err != nil {
		listener.onEvent(Event.NewErrorNoOption(
			Event.ServiceStartFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.ServiceStarting,
			},
		))
		listener.ipRateLimiter.Close()
		listener.ipRateLimiter = nil
		listener.status = Status.Stopped
		return err
	}

	listener.status = Status.Started
	listener.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"tcpSystemgeListener started",
		Event.Context{
			Event.Circumstance: Event.ServiceStarting,
		},
	))
	return nil
}
