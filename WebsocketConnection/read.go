package WebsocketConnection

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Helpers"
)

func (client *WebsocketConnection) read() ([]byte, error) {
	_, messageBytes, err := client.websocketConn.ReadMessage()
	if err != nil {
		if Helpers.IsWebsocketConnClosedErr(err) {
			client.Close()
		}
		return nil, err
	}
	client.BytesReceived.Add(uint64(len(messageBytes)))
	client.MessagesReceived.Add(1)
	return messageBytes, nil
}

func (client *WebsocketConnection) Read() ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine != nil {
		return nil, errors.New("receptionHandler is already running")
	}

	return client.read()
}

func (client *WebsocketConnection) ReadTimeout(timeoutMs uint64) ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	client.SetReadDeadline(timeoutMs)
	return client.read()
}

// can be used to cancel an ongoing read operation
func (client *WebsocketConnection) SetReadDeadline(timeoutMs uint64) {
	client.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}

func (client *WebsocketConnection) SetReadLimit(maxBytes int64) { // i do not comprehend why this is a int64 rather than a uint64
	client.websocketConn.SetReadLimit(maxBytes)
}
