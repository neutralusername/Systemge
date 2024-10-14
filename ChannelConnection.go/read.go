package ChannelConnection

import (
	"errors"
)

func (connection *ChannelConnection[T]) read() (T, error) {
	return <-connection.receiveChannel, nil
}

func (connection *ChannelConnection[T]) Read() (T, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if connection.readRoutine != nil {
		var nilValue T
		return nilValue, errors.New("receptionHandler is already running")
	}

	return connection.read()

}

func (client *ChannelConnection[T]) ReadTimeout(timeoutMs uint64) (T, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	client.SetReadDeadline(timeoutMs)
	return client.read()
}

// can be used to cancel an ongoing read operation
func (connection *ChannelConnection[T]) SetReadDeadline(timeoutMs uint64) {
	/* client.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)) */
}

func (connection *ChannelConnection[T]) SetReadLimit(maxBytes int64) {
	/* 	client.websocketConn.SetReadLimit(maxBytes) */
}
