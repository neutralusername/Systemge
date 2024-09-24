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

	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)

	if event := server.onEvent(Event.NewInfo(
		Event.StartingService,
		"service websocketServer starting",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.StartRoutine,
		}),
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Stoped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"service websocketServer already started",
			server.GetServerContext().Merge(Event.Context{
				Event.Circumstance: Event.StartRoutine,
			}),
		))
		return errors.New("failed to start websocketServer")
	}
	server.status = Status.Pending

	if server.config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(server.config.IpRateLimiter)
	}
	if err := server.httpServer.Start(); err != nil { // TODO: context from this service missing - handle this somehow
		if server.ipRateLimiter != nil {
			server.ipRateLimiter.Close()
			server.ipRateLimiter = nil
		}
		server.status = Status.Stoped
		return err
	}

	server.stopChannel = make(chan bool)

	server.waitGroup.Add(1)
	go server.receiveWebsocketConnectionLoop()

	server.status = Status.Started

	if event := server.onEvent(Event.NewInfo(
		Event.ServiceStarted,
		"service websocketServer started",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.StartRoutine,
		}),
	)); !event.IsInfo() {
		if err := server.stop(false); err != nil {
			panic(err)
		}
	}
	return nil
}
