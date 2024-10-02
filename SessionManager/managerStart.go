package SessionManager

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

func (manager *Manager) Start() error {
	manager.statusMutex.Lock()
	defer manager.statusMutex.Unlock()

	manager.sessionMutex.Lock()
	if manager.isStarted {
		manager.sessionMutex.Unlock()
		return errors.New("session manager already accepting sessions")
	}

	manager.sessionId = Tools.GenerateRandomString(manager.config.SessionIdLength, manager.config.SessionIdAlphabet)
	manager.isStarted = true
	manager.sessionMutex.Unlock()

	return nil
}
