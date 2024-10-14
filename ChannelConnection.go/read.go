package ChannelConnection

import (
	"errors"
)

func (client *ChannelConnection[T]) read() (T, error) {
	return <-client.receiveChannel, nil
}

func (client *ChannelConnection[T]) Read() (T, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine != nil {
		var nilValue T
		return nilValue, errors.New("receptionHandler is already running")
	}

	return client.read()

}

func (client *ChannelConnection[T]) ReadTimeout(timeoutMs uint64) (T, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	client.SetReadDeadline(timeoutMs)
	return client.read()
}

// can be used to cancel an ongoing read operation
func (client *ChannelConnection[T]) SetReadDeadline(timeoutMs uint64) {
	/* client.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)) */
}

func (client *ChannelConnection[T]) SetReadLimit(maxBytes int64) {
	/* 	client.websocketConn.SetReadLimit(maxBytes) */
}
