package SessionManager

import (
	"errors"
)

func (manager *Manager) Stop() error {
	manager.statusMutex.Lock()
	defer manager.statusMutex.Unlock()

	manager.sessionMutex.Lock()
	if !manager.isStarted {
		manager.sessionMutex.Unlock()
		return errors.New("session manager already rejecting sessions")
	}

	manager.isStarted = false
	manager.sessionMutex.Unlock()
	manager.waitgroup.Wait()

	return nil
}
