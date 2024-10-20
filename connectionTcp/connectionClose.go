package connectionTcp

import "errors"

func (connection *TcpConnection) Close() error {
	if !connection.closedMutex.TryLock() {
		return errors.New("connection already closing")
	}
	defer connection.closedMutex.Unlock()

	if connection.closed {
		return errors.New("connection already closed")
	}

	connection.closed = true
	connection.netConn.Close()
	close(connection.closeChannel)

	return nil
}
