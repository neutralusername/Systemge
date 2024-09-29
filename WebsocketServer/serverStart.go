package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *WebsocketServer) Start() error {
	return server.start(true)
}
func (server *WebsocketServer) start(lock bool) error {
	if lock {
		server.statusMutex.Lock()
		defer server.statusMutex.Unlock()
	}

	if event := server.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"service websocketServer starting",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Stopped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"service websocketServer already started",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
		return errors.New("failed to start websocketServer")
	}
	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	server.status = Status.Pending

	if server.config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(server.config.IpRateLimiter)
	}
	if err := server.httpServer.Start(); err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.ServiceStartFailed,
			"failed to start websocketServer's http server",
			Event.Context{
				Event.Circumstance:      Event.ServiceStart,
				Event.TargetServiceType: Event.HttpServer,
			},
		))
		if server.ipRateLimiter != nil {
			server.ipRateLimiter.Close()
			server.ipRateLimiter = nil
		}
		server.status = Status.Stopped
		return err
	}

	server.stopChannel = make(chan bool)

	server.waitGroup.Add(1)
	go server.acceptRoutine()
	server.status = Status.Started

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"service websocketServer started",
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	))
	return nil
}
