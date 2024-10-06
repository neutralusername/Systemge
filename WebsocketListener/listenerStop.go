package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Status"
)

func (listener *WebsocketListener) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status == Status.Stopped {
		return errors.New("websocketListener is already stopped")
	}

	listener.status = Status.Pending
	close(listener.stopChannel)

	if err := listener.httpServer.Stop(); err != nil {
		listener.status = Status.Started
	}

	listener.status = Status.Stopped
	return nil
}
