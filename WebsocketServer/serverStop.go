package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *WebsocketServer) Stop() error {
	return server.stop(true)
}
func (server *WebsocketServer) stop(lock bool) error {
	if lock {
		server.statusMutex.Lock()
		defer server.statusMutex.Unlock()
	}

	if event := server.onEvent(Event.NewInfo(
		Event.StoppingService,
		"service websocketServer stopping",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status == Status.Stopped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"service websocketServer already stopped",
			Event.Context{
				Event.Circumstance: Event.Stop,
			},
		))
		return errors.New("websocketServer not started")
	}
	server.status = Status.Pending

	server.httpServer.Stop()
	if server.ipRateLimiter != nil {
		server.ipRateLimiter.Close()
		server.ipRateLimiter = nil
	}

	close(server.stopChannel)
	server.waitGroup.Wait()
	server.status = Status.Stopped

	event := server.onEvent(Event.NewInfo(
		Event.ServiceStopped,
		"service websocketServer stopped",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	))
	if !event.IsInfo() {
		if err := server.start(false); err != nil {
			panic(err)
		}
	}
	return nil
}
