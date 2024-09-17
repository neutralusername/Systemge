package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpSystemgeConnection) send(bytes []byte) error {
	connection.sendMutex.Lock()
	defer connection.sendMutex.Unlock()
	bytesSent, err := Tcp.Send(connection.netConn, bytes, connection.config.TcpSendTimeoutMs)
	if err != nil {
		if Tcp.IsConnectionClosed(err) {
			connection.Close()
			return Error.New("Connection closed", err)
		}
		return err
	}
	connection.bytesSent.Add(bytesSent)
	return nil
}
