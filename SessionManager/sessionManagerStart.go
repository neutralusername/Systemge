package SessionManager

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

func (manager *Manager) Start() error {
	manager.statusMutex.Lock()
	defer manager.statusMutex.Unlock()

	if event := manager.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"starting sessionManager",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	manager.sessionMutex.Lock()
	if manager.isStarted {
		manager.sessionMutex.Unlock()
		manager.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"session manager already started",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
		return errors.New("session manager already accepting sessions")
	}

	manager.sessionId = Tools.GenerateRandomString(manager.config.SessionIdLength, manager.config.SessionIdAlphabet)
	manager.isStarted = true
	manager.sessionMutex.Unlock()

	manager.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"sessionManager started",
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	))
	return nil
}
