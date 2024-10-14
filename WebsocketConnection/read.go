package WebsocketConnection

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Helpers"
)

func (connection *WebsocketConnection) read() ([]byte, error) {
	_, messageBytes, err := connection.websocketConn.ReadMessage()
	if err != nil {
		if Helpers.IsWebsocketConnClosedErr(err) {
			connection.Close()
		}
		return nil, err
	}
	connection.BytesReceived.Add(uint64(len(messageBytes)))
	connection.MessagesReceived.Add(1)
	return messageBytes, nil
}

func (connection *WebsocketConnection) Read() ([]byte, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if connection.readRoutine != nil {
		return nil, errors.New("receptionHandler is already running")
	}

	return connection.read()
}

func (connection *WebsocketConnection) ReadTimeout(timeoutMs uint64) ([]byte, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if connection.readRoutine != nil {
		return nil, errors.New("receptionHandler is already running")
	}

	connection.SetReadDeadline(timeoutMs)
	return connection.read()
}

// can be used to cancel an ongoing read operation
func (connection *WebsocketConnection) SetReadDeadline(timeoutMs uint64) {
	connection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}

func (connection *WebsocketConnection) SetReadLimit(maxBytes int64) { // i do not comprehend why this is a int64 rather than a uint64
	connection.websocketConn.SetReadLimit(maxBytes)
}
