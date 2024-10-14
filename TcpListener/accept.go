package TcpListener

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/TcpConnection"
)

func (listener *TcpListener) AcceptConnection(connectionConfig *Config.TcpSystemgeConnection) (*TcpConnection.TcpConnection, error) {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	netConn, err := listener.tcpListener.Accept()
	if err != nil {
		listener.ClientsFailed.Add(1)
		return nil, err
	}

	tcpSystemgeConnection, err := TcpConnection.New(connectionConfig, netConn)
	if err != nil {
		listener.ClientsFailed.Add(1)
		netConn.Close()
		return nil, err
	}

	return tcpSystemgeConnection, nil
}
