package ListenerTcp

import (
	"github.com/neutralusername/Systemge/ConnectionTcp"
)

func (listener *TcpListener) Accept( /* connectionConfig *Config.TcpSystemgeConnection */ timeoutNs int64) (*ConnectionTcp.TcpConnection, error) {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	// net.Listener does not implement a timeout for Accept... lf solution
	netConn, err := listener.tcpListener.Accept()
	if err != nil {
		listener.ClientsFailed.Add(1)
		return nil, err
	}

	tcpSystemgeConnection, err := ConnectionTcp.New( /* connectionConfig, */ netConn)
	if err != nil {
		listener.ClientsFailed.Add(1)
		netConn.Close()
		return nil, err
	}

	listener.ClientsAccepted.Add(1)
	return tcpSystemgeConnection, nil
}
