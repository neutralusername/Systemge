package connectionTcp

import (
	"time"

	"github.com/neutralusername/systemge/helpers"
)

func (client *TcpConnection) ReadChannel(stopChannel <-chan struct{}) <-chan []byte {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	resultChannel := make(chan []byte)
	doneChannel := make(chan struct{})
	go func() {
		defer func() {
			close(resultChannel)
			close(doneChannel)
		}()

		bytes, err := client.Read(0)
		if err != nil {
			return
		}
		resultChannel <- bytes
	}()

	go func() {
		select {
		case <-stopChannel:
			client.SetReadDeadline(1)
			return
		case <-doneChannel:
			return
		}
	}()

	return resultChannel
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
