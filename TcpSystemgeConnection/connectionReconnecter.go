package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type Reconnector struct {
	abortChannel       chan bool
	systemgeConnection SystemgeConnection.SystemgeConnection
}

func StartReconnector(systemgeConnection SystemgeConnection.SystemgeConnection, config *Config.SystemgeConnectionAttempt) *Reconnector {
	reconnector := &Reconnector{
		abortChannel:       make(chan bool),
		systemgeConnection: systemgeConnection,
	}
	go reconnector.startReconnects()

	return reconnector
}

func (reconnector *Reconnector) startReconnects() {
	select {
	case <-reconnector.abortChannel:

	case <-reconnector.systemgeConnection.GetCloseChannel():

	}
}
