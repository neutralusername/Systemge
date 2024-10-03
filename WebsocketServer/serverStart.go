package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (server *WebsocketServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if server.eventHandler != nil {
		if event := server.onEvent(Event.New(
			Event.ServiceStarting,
			Event.ServiceStart,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return errors.New("failed to start websocketServer")
		}
	}

	if server.status != Status.Stopped {
		if server.eventHandler != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.ServiceAlreadyStarted,
				"service websocketServer already started",
				Event.Context{
					Event.Circumstance: Event.ServiceStart,
				},
			))
		}
		return errors.New("failed to start websocketServer")
	}
	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	server.status = Status.Pending

	server.websocketListener.Start()
	server.stopChannel = make(chan bool)

	server.waitGroup.Add(1)
	go server.sessionRoutine()
	server.status = Status.Started

	if server.eventHandler != nil {
		server.onEvent(Event.NewInfoNoOption(
			Event.ServiceStarted,
			"service websocketServer started",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
	}
	return nil
}
