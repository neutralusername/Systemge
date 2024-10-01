package SessionManager

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
)

func (manager *Manager) Stop() error {
	manager.statusMutex.Lock()
	defer manager.statusMutex.Unlock()

	if event := manager.onEvent(Event.NewInfo(
		Event.ServiceStopping,
		"stopping sessionManager",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	manager.sessionMutex.Lock()
	if !manager.isStarted {
		manager.sessionMutex.Unlock()
		manager.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"session manager already stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		))
		return errors.New("session manager already rejecting sessions")
	}

	manager.isStarted = false
	manager.sessionMutex.Unlock()
	manager.waitgroup.Wait()

	manager.onEvent(Event.NewInfoNoOption(
		Event.ServiceStopped,
		"sessionManager stopped",
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	))

	return nil
}
