package listenerTcp

import (
	"errors"
	"net"
	"time"

	"github.com/neutralusername/systemge/connectionTcp"
	"github.com/neutralusername/systemge/systemge"
)

func (listener *TcpListener) AcceptChannel() <-chan systemge.Connection[[]byte] {
	resultChannel := make(chan systemge.Connection[[]byte])
	go func() {
		defer close(resultChannel)

		connectionRequest, err := listener.Accept(0)
		if err != nil {
			return
		}
		resultChannel <- connectionRequest
	}()

	return resultChannel
}

func (listener *TcpListener) Accept(timeoutNs int64) (systemge.Connection[[]byte], error) {
	listener.SetAcceptTimeout(timeoutNs)

	l := listener.tcpListener
	if l == nil {
		return nil, errors.New("tcpSystemgeListener is not started")
	}

	if tlsListener := listener.tlsListener; tlsListener != nil {
		l = tlsListener
	}

	netConn, err := l.Accept()
	if err != nil {
		listener.ClientsFailed.Add(1)
		return nil, err
	}

	tcpSystemgeConnection, err := connectionTcp.New(listener.tcpBufferedReaderConfig, netConn)
	if err != nil {
		listener.ClientsFailed.Add(1)
		netConn.Close()
		return nil, err
	}

	listener.ClientsAccepted.Add(1)
	return tcpSystemgeConnection, nil
}

func (listener *TcpListener) SetAcceptTimeout(timeoutNs int64) {
	l := listener.tcpListener
	if l == nil {
		return
	}
	tcpListener := l.(*net.TCPListener)
	tcpListener.SetDeadline(time.Now().Add(time.Duration(timeoutNs)))
}
