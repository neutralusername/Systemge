package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpConnection) heartbeatLoop() {
	for {
		select {
		case <-connection.closeChannel:
			return
		default:
			connection.sendMutex.Lock()
			err := Tcp.SendHeartbeat(connection.netConn, connection.config.TcpSendTimeoutMs)
			connection.sendMutex.Unlock()
			if err != nil {
				if Tcp.IsConnectionClosed(err) {
					connection.Close()
					return
				}
				continue
			}
			connection.bytesSent.Add(1)
			time.Sleep(time.Duration(connection.config.HeartbeatIntervalMs) * time.Millisecond)
		}
	}
}
