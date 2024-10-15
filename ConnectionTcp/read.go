package ConnectionTcp

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Helpers"
)

func (client *TcpConnection) Read(timeoutNs int64) ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine != nil {
		return nil, errors.New("receptionHandler is already running")
	}

	client.SetReadDeadline(timeoutNs)
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

func (client *TcpConnection) SetReadDeadline(timeoutNs int64) {
	client.netConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
}
