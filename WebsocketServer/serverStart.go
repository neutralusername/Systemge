package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketListener"
)

func (server *WebsocketServer) Start() error {
	return server.start(true)
}
func (server *WebsocketServer) start(lock bool) error {
	if lock {
		server.statusMutex.Lock()
		defer server.statusMutex.Unlock()
	}

	if event := server.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"service websocketServer starting",
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
			"service websocketServer already started",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
		return errors.New("failed to start websocketServer")
	}
	server.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	server.status = Status.Pending

	websocketListener, err := WebsocketListener.New(server.name+"_websocketListener", server.config.WebsocketListenerConfig, server.whitelist, server.blacklist, server.eventHandler)
	if err != nil {
		return err
	}
	server.websocketListener = websocketListener
	server.stopChannel = make(chan bool)

	server.waitGroup.Add(1)
	go server.sessionRoutine()
	server.status = Status.Started

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"service websocketServer started",
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	))
	return nil
}
