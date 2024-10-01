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

	if event := server.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"starting systemgeServer",
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
			"systemgeServer not stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
		return errors.New("failed to start systemge server")
	}
	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	server.status = Status.Pending

	listener, err := TcpSystemgeListener.New(server.name, server.config.TcpSystemgeListenerConfig, server.whitelist, server.blacklist, server.eventHandler)
	if err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.ServiceStartFailed,
			"failed to initialize tcp systemge listener",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
		server.status = Status.Stopped
		return err
	}
	server.listener = listener
	server.stopChannel = make(chan bool)

	server.sessionManager.Start()

	server.waitGroup.Add(1)
	go server.connectionRoutine()

	server.status = Status.Started

	server.onEvent(Event.NewInfo(
		Event.ServiceStarted,
		"systemgeServer started",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	))
	return nil
}
