package Oauth2Server

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *Server) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.ServiceStopping,
		"Stopping oauth2 server",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	)); event != nil {
		return event.GetError()
	}

	if server.status != Status.Started {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"oauth2 server already stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		))
		return errors.New("oauth2 server is already stopped")
	}
	server.status = Status.Pending
	err := server.httpServer.Stop()
	if err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.ServiceStopFailed,
			"Failed to stop oauth2' http server",
			Event.Context{
				Event.Circumstance:      Event.ServiceStop,
				Event.TargetServiceType: Event.HttpServer,
			},
		))
	}

	server.status = Status.Stopped

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStopped,
		"oauth2 server stopped",
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	))

	return nil
}
