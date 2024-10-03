package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (server *WebsocketServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if server.eventHandler != nil {
		if event := server.onEvent(Event.NewInfo(
			Event.ServiceStopping,
			"service websocketServer stopping",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		)); !event.IsInfo() {
			return event.GetError()
		}
	}

	if server.status == Status.Stopped {
		if server.eventHandler != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.ServiceAlreadyStopped,
				"service websocketServer already stopped",
				Event.Context{
					Event.Circumstance: Event.ServiceStop,
				},
			))
		}
		return errors.New("websocketServer not started")
	}
	server.status = Status.Pending

	server.websocketListener.Stop()

	close(server.stopChannel)
	server.waitGroup.Wait()
	server.status = Status.Stopped

	if server.eventHandler != nil {
		server.onEvent(Event.NewInfoNoOption(
			Event.ServiceStopped,
			"service websocketServer stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		))
	}
	return nil
}
