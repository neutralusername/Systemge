package WebsocketClient

import (
	"errors"
	"time"
)

func (client *WebsocketClient) read() ([]byte, error) {
	_, messageBytes, err := client.websocketConn.ReadMessage()
	if err != nil {
		if isWebsocketConnClosedErr(err) {
			client.Close()
		}
		return nil, err
	}
	client.BytesReceived.Add(uint64(len(messageBytes)))
	client.MessagesReceived.Add(1)
	return messageBytes, nil
}

func (client *WebsocketClient) Read() ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readHandler != nil {
		return nil, errors.New("receptionHandler is already running")
	}

	return client.read()
}

// can be used to cancel an ongoing read operation
func (client *WebsocketClient) SetReadDeadline(timeoutMs uint64) {
	client.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}
