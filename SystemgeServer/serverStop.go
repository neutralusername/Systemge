package SystemgeServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()

	if event := server.onEvent(Event.NewInfo(
		Event.ServiceStoping,
		"stopping systemgeServer",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if server.status != Status.Started {
		server.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStoped,
			"systemgeServer is already stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		))
		return errors.New("server is already stopped")
	}

	server.status = Status.Pending
	close(server.stopChannel)
	if err := server.listener.Close(); err != nil {
		server.onEvent(Event.NewErrorNoOption(
			Event.ServiceStopFailed,
			"failed to close listener",
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		))
	}
	server.sessionManager.Stop()
	for _, identity := range server.sessionManager.GetIdentities() {
		for _, session := range server.sessionManager.GetSessions(identity) {
			connection, ok := session.Get("connection")
			if !ok {
				continue
			}
			connection.(SystemgeConnection.SystemgeConnection).Close()
		}
	}
	server.waitGroup.Wait()
	server.status = Status.Stoped

	server.onEvent(Event.NewInfoNoOption(
		Event.ServiceStoped,
		"systemgeServer stopped",
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	))
	return nil
}
