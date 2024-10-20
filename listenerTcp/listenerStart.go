package listenerTcp

import (
	"errors"

	"github.com/neutralusername/systemge/constants"
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

	listener.status = status.Started
	return nil
}
