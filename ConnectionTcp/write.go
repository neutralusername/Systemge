package ConnectionTcp

import (
	"time"

	"github.com/neutralusername/Systemge/Helpers"
)

const ENDOFMESSAGE = '\x04'
const HEARTBEAT = '\x05'

func (connection *TcpConnection) SendHeartbeat(timeoutNs int64) error {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if timeoutNs > 0 {
		connection.netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
	} else {
		connection.netConn.SetWriteDeadline(time.Time{})
	}
	_, err := connection.netConn.Write([]byte{HEARTBEAT})
	if err != nil {
		return err
	}
	connection.BytesSent.Add(1)
	connection.MessagesSent.Add(1)
	return nil
}

func (client *TcpConnection) Write(messageBytes []byte, timeoutNs int64) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	client.SetWriteDeadline(timeoutNs)
	_, err := client.netConn.Write(append(messageBytes, ENDOFMESSAGE))
	if err != nil {
		if Helpers.IsNetConnClosedErr(err) {
			client.Close()
		}
		return err
	}
	client.BytesSent.Add(uint64(len(messageBytes)))
	client.MessagesSent.Add(1)
	return nil
}

func (client *TcpConnection) SetWriteDeadline(timeoutNs int64) {
	client.netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutNs) * time.Nanosecond))
}
