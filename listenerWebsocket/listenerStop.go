package listenerWebsocket

import (
	"errors"

	"github.com/neutralusername/Systemge/status"
)

func (listener *WebsocketListener) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status == status.Stopped {
		return errors.New("websocketListener is already stopped")
	}

	listener.status = status.Pending
	close(listener.stopChannel)

	if err := listener.httpServer.Stop(); err != nil {
		// something
	}

	listener.status = status.Stopped
	return nil
}
