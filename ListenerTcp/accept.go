package ListenerTcp

import (
	"net"
	"time"

	"github.com/neutralusername/Systemge/ConnectionTcp"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (listener *TcpListener) Accept(timeoutNs int64) (SystemgeConnection.SystemgeConnection[[]byte], error) {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	l := listener.tcpListener
	tcpListener := l.(*net.TCPListener)
	tcpListener.SetDeadline(time.Now().Add(time.Duration(timeoutNs)))

	if listener.tlsListener != nil {
		l = listener.tlsListener
	}

	netConn, err := l.Accept()
	if err != nil {
		listener.ClientsFailed.Add(1)
		return nil, err
	}

	tcpSystemgeConnection, err := ConnectionTcp.New(netConn)
	if err != nil {
		listener.ClientsFailed.Add(1)
		netConn.Close()
		return nil, err
	}

	listener.ClientsAccepted.Add(1)
	return tcpSystemgeConnection, nil
}
