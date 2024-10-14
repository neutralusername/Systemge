package TcpConnection

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Helpers"
)

func (client *TcpConnection) read() ([]byte, error) {
	messageBytes, err := client.messageReceiver.ReadNextMessage()
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

func (client *TcpConnection) Read() ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine != nil {
		return nil, errors.New("receptionHandler is already running")
	}

	return client.read()
}

func (client *TcpConnection) ReadTimeout(timeoutMs uint64) ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	client.SetReadDeadline(timeoutMs)
	return client.read()
}

// can be used to cancel an ongoing read operation
func (client *TcpConnection) SetReadDeadline(timeoutMs uint64) {
	client.netConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}
