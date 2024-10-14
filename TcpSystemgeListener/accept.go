package TcpSystemgeListener

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (listener *TcpSystemgeListener) AcceptConnection(connectionConfig *Config.TcpSystemgeConnection) (SystemgeConnection.SystemgeConnection, error) {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	netConn, err := listener.tcpListener.Accept()
	if err != nil {
		listener.tcpSystemgeConnectionAttemptsFailed.Add(1)
		return nil, err
	}

	connection, err := listener.serverHandshake(connectionConfig, netConn, eventHandler)
	if err != nil {
		listener.tcpSystemgeConnectionAttemptsRejected.Add(1)
		netConn.Close()
		return nil, err
	}

	listener.tcpSystemgeConnectionAttemptsAccepted.Add(1)
	return connection, nil
}
