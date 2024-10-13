package WebsocketClient

import (
	"errors"
)

func (client *WebsocketClient) Close() error {
	if !client.closedMutex.TryLock() {
		return errors.New("websocketClient already closing")
	}
	defer client.closedMutex.Unlock()

	if client.closed {
		return errors.New("websocketClient already closed")
	}

	client.closed = true
	client.websocketConn.Close()
	close(client.closeChannel)

	if client.readRoutine != nil {
		client.readRoutine.StopRoutine()
	}

	return nil
}
