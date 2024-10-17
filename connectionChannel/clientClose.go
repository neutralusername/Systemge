package connectionChannel

import (
	"errors"
)

func (connection *ChannelConnection[T]) Close() error {
	if !connection.closedMutex.TryLock() {
		return errors.New("websocketClient already closing")
	}
	defer connection.closedMutex.Unlock()

	if connection.closed {
		return errors.New("websocketClient already closed")
	}

	connection.closed = true

	close(connection.closeChannel)

	return nil
}
