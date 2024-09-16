package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpConnection) receive() ([]byte, error) {
	connection.receiveMutex.Lock()
	defer connection.receiveMutex.Unlock()
	bytes, err := connection.messageReceiver.ReceiveNextMessage()
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (connection *TcpConnection) send(bytes []byte) error {
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
