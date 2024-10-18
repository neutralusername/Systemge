package connectionTcp

import (
	"time"

	"github.com/neutralusername/systemge/helpers"
)

func (client *TcpConnection) ReadChannel(stopChannel <-chan struct{}) ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	doneChannel := make(chan struct{})
	go func() {
		select {
		case <-stopChannel:
			client.SetReadDeadline(1)
			return
		case <-doneChannel:
			return
		}
	}()

	bytes, err := client.Read(0)
	close(doneChannel)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (client *TcpConnection) Read(timeoutNs int64) ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	client.SetReadDeadline(timeoutNs)
	messageBytes, newBytesRead, err := client.tcpBufferedReader.Read()
	if err != nil {
		if helpers.IsNetConnClosedErr(err) {
			client.Close()
		}
		return nil, err
	}
	client.BytesReceived.Add(uint64(newBytesRead))
	client.MessagesReceived.Add(1)
	return messageBytes, nil
}

func (client *TcpConnection) SetReadDeadline(timeoutNs int64) {
	client.netConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
}
