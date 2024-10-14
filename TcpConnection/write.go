package TcpConnection

import (
	"errors"
	"net"
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

func SendHeartbeat(netConn net.Conn, timeoutMs uint64) error {
	if netConn == nil {
		return errors.New("net.Conn is nil")
	}

	return nil
}

func (client *TcpConnection) write(messageBytes []byte) error {
	_, err := client.netConn.Write(append(messageBytes, ENDOFMESSAGE))
	if err != nil {
		if Helpers.IsConnectionClosed(err) {
			client.Close()
		}
		return err
	}
	client.BytesSent.Add(uint64(len(messageBytes)))
	client.MessagesSent.Add(1)
	return nil
}

func (client *TcpConnection) Write(messageBytes []byte) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	return client.write(messageBytes)
}

func (client *TcpConnection) WriteTimeout(messageBytes []byte, timeoutMs uint64) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	client.SetWriteDeadline(timeoutMs)
	return client.write(messageBytes)
}

func (client *TcpConnection) SetWriteDeadline(timeoutMs uint64) {
	client.netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}
