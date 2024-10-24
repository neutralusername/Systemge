package listenerTcp

import (
	"crypto/tls"
	"errors"
	"net"

	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

func (listener *TcpListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status != status.Stopped {
		return errors.New("tcpSystemgeListener is already started")
	}

	listener.sessionId = tools.GenerateRandomString(constants.SessionIdLength, tools.ALPHA_NUMERIC)

	if listener.config.Ip == "" {
		ip, err := net.LookupIP(listener.config.Domain)
		if err != nil {
			return err
		}
		listener.config.Ip = ip[0].String()
	}

	tcpListener, err := NewTcpListener(listener.config.Ip + ":" + helpers.Uint16ToString(listener.config.Port))
	if err != nil {
		return err
	}
	listener.tcpListener = tcpListener

	if listener.config.TlsCertPath != "" && listener.config.TlsKeyPath != "" {
		tlsListener, err := NewTlsListener(tcpListener, listener.config.TlsCertPath, listener.config.TlsKeyPath)
		if err != nil {
			tcpListener.Close()
			return err
		}
		listener.tlsListener = tlsListener
	}

	listener.stopChannel = make(chan struct{})

	listener.status = status.Started
	return nil
}

func NewTcpListener(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func NewTlsListener(listener net.Listener, tlsCertPath, tlsKeyPath string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return tls.NewListener(listener, tlsConfig), nil
}
