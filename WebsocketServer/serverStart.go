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

	event := server.onEvent(Event.New(
		Starting,
		Event.Context{},
		Event.Continue,
		Event.Cancel,
	))
	if event.GetAction() == Event.Cancel {
		return errors.New(Starting)
	}

	if server.status != Status.Stopped {
		server.onEvent(Event.New(
			AlreadyStarted,
			Event.Context{},
			Event.Cancel,
		))
		return errors.New(AlreadyStarted)
	}
	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	server.status = Status.Pending

	server.websocketListener.Start()
	server.stopChannel = make(chan bool)

	server.waitGroup.Add(1)
	go server.sessionRoutine()
	server.status = Status.Started

	server.onEvent(Event.New(
		Started,
		Event.Context{},
		Event.Continue,
	))

	return nil
}
