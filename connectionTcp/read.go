package connectionTcp

import (
	"time"

	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/tools"
)

func (client *TcpConnection) ReadChannel() <-chan []byte {
	return tools.ChannelCall(func() ([]byte, error) {
		return client.Read(0)
	})
}

func (client *TcpConnection) Read(timeoutNs int64) ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	client.SetReadDeadline(timeoutNs)
	data, newBytesRead, err := client.tcpBufferedReader.Read()
	if err != nil {
		if helpers.IsNetConnClosedErr(err) {
			client.Close()
		}
		return nil, err
	}
	client.BytesReceived.Add(uint64(newBytesRead))
	client.MessagesReceived.Add(1)
	return data, nil
}

func (client *TcpConnection) SetReadDeadline(timeoutNs int64) {
	client.netConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
}
