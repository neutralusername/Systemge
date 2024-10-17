package listenerTcp

import (
	"errors"

	"github.com/neutralusername/Systemge/status"
)

// closing this will not automatically close all connections accepted by this listener. use SystemgeServer if this functionality is desired.
func (listener *TcpListener) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status != status.Started {
		return errors.New("tcpSystemgeListener is already stopped")
	}

	listener.tcpListener.Close()

	if listener.tlsListener != nil {
		listener.tlsListener.Close()
	}

	listener.status = status.Stopped
	return nil
}
