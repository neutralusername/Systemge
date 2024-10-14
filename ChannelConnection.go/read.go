package ChannelConnection

import (
	"errors"
)

func (client *ChannelConnection[T]) read() (T, error) {
	/* _, messageBytes, err := client.websocketConn.ReadMessage()
	if err != nil {
		if Helpers.IsWebsocketConnClosedErr(err) {
			client.Close()
		}
		return nil, err
	} */
	client.BytesReceived.Add(uint64(len(messageBytes)))
	client.MessagesReceived.Add(1)
	return messageBytes, nil
}

func (client *ChannelConnection[T]) Read() (T, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine != nil {
		return nil, errors.New("receptionHandler is already running")
	}
	/*
		return client.read()
	*/
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
