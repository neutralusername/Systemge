package listenerTcp

import (
	"net"
	"time"

	"github.com/neutralusername/Systemge/connectionTcp"
	"github.com/neutralusername/Systemge/systemge"
)

func (listener *TcpListener) Accept(timeoutNs int64) (systemge.Connection[[]byte], error) {

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

	tcpSystemgeConnection, err := connectionTcp.New(listener.connectionConfig, netConn)
	if err != nil {
		listener.ClientsFailed.Add(1)
		netConn.Close()
		return nil, err
	}

	listener.ClientsAccepted.Add(1)
	return tcpSystemgeConnection, nil
}
