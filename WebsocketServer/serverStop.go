package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *WebsocketServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	event := server.onEvent(Event.New(
		Stoping,
		Event.Context{},
		Event.Continue,
		Event.Cancel,
	))
	if event.GetAction() == Event.Cancel {
		return errors.New(Stoping)
	}

	if server.status == Status.Stopped {
		server.onEvent(Event.New(
			AlreadyStoped,
			Event.Context{},
			Event.Cancel,
		))
		return errors.New(AlreadyStoped)
	}
	server.status = Status.Pending

	server.websocketListener.Stop()

	close(server.stopChannel)
	server.waitGroup.Wait()
	server.status = Status.Stopped

	server.onEvent(Event.New(
		Stoped,
		Event.Context{},
		Event.Continue,
	))
	return nil
}
