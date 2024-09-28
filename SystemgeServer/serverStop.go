package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *SystemgeServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.StoppingService,
		"stopping server",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Started {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"server is already stopped",
			Event.Context{
				Event.Circumstance: Event.Stop,
			},
		))
		return errors.New("server is already stopped")
	}

	server.status = Status.Pending
	close(server.stopChannel)
	err := server.listener.Close()
	if err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.CloseFailed,
			"failed to close listener",
			Event.Context{
				Event.Circumstance: Event.Stop,
			},
		))
	}
	server.waitGroup.Wait()
	server.stopChannel = nil
	server.listener = nil

	server.status = Status.Stopped

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStopped,
		"server stopped",
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	))
	return nil
}
