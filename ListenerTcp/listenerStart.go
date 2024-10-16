package ListenerTcp

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *TcpListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status != Status.Stopped {
		return errors.New("tcpSystemgeListener is already started")
	}

	listener.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	tcpListener, err := NewTcpListener(listener.config.TcpServerConfig.Port)
	if err != nil {
		return err
	}
	listener.tcpListener = tcpListener

	if listener.config.TcpServerConfig.TlsCertPath != "" && listener.config.TcpServerConfig.TlsKeyPath != "" {
		tlsListener, err := NewTlsListener(tcpListener, listener.config.TcpServerConfig.TlsCertPath, listener.config.TcpServerConfig.TlsKeyPath)
		if err != nil {
			tcpListener.Close()
			return err
		}
		listener.tlsListener = tlsListener
	}

	listener.stopChannel = make(chan struct{})

	listener.status = Status.Started
	return nil
}
