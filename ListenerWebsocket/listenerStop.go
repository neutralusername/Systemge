package ListenerWebsocket

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
		// something
	}
	listener.StopAcceptRoutine(false)

	listener.status = Status.Stopped
	return nil
}