package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/TcpSystemgeListener"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *SystemgeServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)

	if event := server.onEvent(Event.NewInfo(
		Event.StartingService,
		"starting systemgeServer",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Start,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Stopped {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"systemgeServer not stopped",
			Event.Context{
				Event.Circumstance: Event.Start,
			},
		))
		return errors.New("failed to start systemge server")
	}
	server.status = Status.Pending

	listener, err := TcpSystemgeListener.New(server.name, server.config.TcpSystemgeListenerConfig, server.whitelist, server.blacklist, server.eventHandler)
	if err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.InitializationFailed,
			"failed to initialize tcp systemge listener",
			Event.Context{
				Event.Circumstance: Event.Start,
			},
		))
		server.status = Status.Stopped
		return err
	}
	server.listener = listener
	server.stopChannel = make(chan bool)

	server.waitGroup.Add(1)
	go server.acceptRoutine()

	server.status = Status.Started

	server.onEvent(Event.NewInfo(
		Event.ServiceStarted,
		"systemgeServer started",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Start,
		},
	))
	return nil
}