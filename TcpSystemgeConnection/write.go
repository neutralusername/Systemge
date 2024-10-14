package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpSystemgeConnection) SendHeartbeat() error {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	err := Tcp.SendHeartbeat(connection.netConn, connection.config.TcpSendTimeoutMs)
	if err != nil {
		if Tcp.IsConnectionClosed(err) {
			connection.Close()
		}
		return err
	}
	connection.BytesSent.Add(1)
	connection.MessagesSent.Add(1)

	return nil
}

func (client *TcpSystemgeConnection) write(messageBytes []byte) error {
	_, err := client.netConn.Write(messageBytes)
	if err != nil {
		if Tcp.IsConnectionClosed(err) {
			client.Close()
		}
		return err
	}
	client.BytesSent.Add(uint64(len(messageBytes)))
	client.MessagesSent.Add(1)
	return nil
}

func (client *TcpSystemgeConnection) Write(messageBytes []byte) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	return client.write(messageBytes)
}

func (client *TcpSystemgeConnection) WriteTimeout(messageBytes []byte, timeoutMs uint64) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	client.SetWriteDeadline(timeoutMs)
	return client.write(messageBytes)
}

func (client *TcpSystemgeConnection) SetWriteDeadline(timeoutMs uint64) {
	client.netConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
}
