package Oauth2Server

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *Server) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"Starting oauth2 server",
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
			"oauth2 server already started",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
		return errors.New("Server is not in stopped state")
	}
	server.sessionId = server.randomizer.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	server.status = Status.Pending

	if err := server.httpServer.Start(); err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.ServiceStartFailed,
			"Failed to start oauth2' http server",
			Event.Context{
				Event.Circumstance:      Event.ServiceStart,
				Event.TargetServiceType: Event.HttpServer,
			},
		))
		server.status = Status.Stopped
		return err
	}

	server.status = Status.Started

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"Oauth2 server started",
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	))

	return nil
}
