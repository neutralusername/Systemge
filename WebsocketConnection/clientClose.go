package WebsocketConnection

import (
	"errors"
)

func (connection *WebsocketConnection) Close() error {
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

	if connection.readRoutine != nil {
		connection.StopReadRoutine()
	}

	return nil
}
