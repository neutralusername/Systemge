package TcpSystemgeConnection

import "github.com/neutralusername/Systemge/Tcp"

func (connection *TcpSystemgeConnection) SendHeartbeat() error {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	return Tcp.SendHeartbeat(connection.netConn, connection.config.TcpSendTimeoutMs)
}
