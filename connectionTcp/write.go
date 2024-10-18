package connectionTcp

import (
	"time"

	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/tools"
)

func (connection *TcpConnection) SendHeartbeat(timeoutNs int64) error {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if timeoutNs > 0 {
		connection.netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
	} else {
		connection.netConn.SetWriteDeadline(time.Time{})
	}
	_, err := connection.netConn.Write([]byte{tools.HEARTBEAT})
	if err != nil {
		return err
	}
	connection.BytesSent.Add(1)
	connection.MessagesSent.Add(1)
	return nil
}

func (client *TcpConnection) Write(data []byte, timeoutNs int64) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	client.SetWriteDeadline(timeoutNs)
	_, err := client.netConn.Write(append(data, tools.ENDOFMESSAGE))
	if err != nil {
		if helpers.IsNetConnClosedErr(err) {
			client.Close()
		}
		return err
	}
	client.BytesSent.Add(uint64(len(data)))
	client.MessagesSent.Add(1)
	return nil
}

func (client *TcpConnection) SetWriteDeadline(timeoutNs int64) {
	client.netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
}
