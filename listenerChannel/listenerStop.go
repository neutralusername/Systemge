package listenerChannel

import (
	"errors"

	"github.com/neutralusername/systemge/status"
)

func (listener *ChannelListener[T]) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status == status.Stopped {
		return errors.New("websocketListener is already stopped")
	}

	listener.status = status.Pending
	close(listener.stopChannel)
	listener.connectionChannel = nil

	listener.status = status.Stopped
	return nil
}
