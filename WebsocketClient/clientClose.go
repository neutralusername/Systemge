package WebsocketClient

import (
	"errors"
)

func (connection *WebsocketClient) Close() error {
	if !connection.closedMutex.TryLock() {
		return errors.New("websocketClient already closing")
	}
	defer connection.closedMutex.Unlock()

	if connection.closed {
		return errors.New("websocketClient already closed")
	}

	connection.closed = true
	connection.websocketConn.Close()
	close(connection.closeChannel)

	return nil
}
