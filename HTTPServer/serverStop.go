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
		Event.ServiceStopping,
		"Stopping http server",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Started {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"http server not started",
			Event.Context{
				Event.Circumstance: Event.Stop,
			},
		))
		return errors.New("http server not started")
	}
	server.status = Status.Pending

	err := server.httpServer.Close()
	if err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.CloseFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.Stop,
			},
		))
	}
	server.httpServer = nil
	server.status = Status.Stopped

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStopped,
		"http server stopped",
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	))
	return nil
}
