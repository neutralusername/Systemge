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
		event := server.onEvent(Event.New(
			Event.ServiceStarting,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return errors.New(Event.ServiceStarting)
		}
	}

	if server.status != Status.Stopped {
		if server.eventHandler != nil {
			server.onEvent(Event.New(
				Event.ServiceAlreadyStarted,
				Event.Context{},
				Event.Cancel,
			))
		}
		return errors.New(Event.ServiceAlreadyStarted)
	}
	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	server.status = Status.Pending

	server.websocketListener.Start()
	server.stopChannel = make(chan bool)

	server.waitGroup.Add(1)
	go server.sessionRoutine()
	server.status = Status.Started

	if server.eventHandler != nil {
		server.onEvent(Event.New(
			Event.ServiceStarted,
			Event.Context{},
			Event.Continue,
		))
	}

	return nil
}
