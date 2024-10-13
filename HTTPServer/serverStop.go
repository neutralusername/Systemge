package HTTPServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *HTTPServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.ServiceStoping,
		"Stopping http server",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Started {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStoped,
			"http server not started",
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		))
		return errors.New("http server not started")
	}
	server.status = Status.Pending

	err := server.httpServer.Close()
	if err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.ServiceStopFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		))
	}
	server.httpServer = nil
	server.status = Status.Stoped

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStoped,
		"http server stopped",
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	))
	return nil
}
