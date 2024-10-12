package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *WebsocketServer[O]) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if server.eventHandler != nil {
		event := server.eventHandler.Handle(Event.New(
			Event.ServiceStoping,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return errors.New(Event.ServiceStoping)
		}
	}

	if server.status == Status.Stopped {
		if server.eventHandler != nil {
			server.eventHandler.Handle(Event.New(
				Event.ServiceAlreadyStoped,
				Event.Context{},
				Event.Cancel,
			))
		}
		return errors.New(Event.ServiceAlreadyStoped)
	}
	server.status = Status.Pending

	server.websocketListener.Stop()

	close(server.stopChannel)
	server.waitGroup.Wait()
	server.status = Status.Stopped

	if server.eventHandler != nil {
		server.eventHandler.Handle(Event.New(
			Event.ServiceStoped,
			Event.Context{},
			Event.Continue,
		))
	}
	return nil
}
