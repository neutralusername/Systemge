package TcpSystemgeConnection

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

type Reconnector struct {
	abortChannel chan bool
}

func StartReconnector(systemgeConnection SystemgeConnection.SystemgeConnection, config *Config.SystemgeConnectionAttempt) *Reconnector {
	reconnector := &Reconnector{}

	go func() {
		select {
		case <-reconnector.abortChannel:

		case <-systemgeConnection.GetCloseChannel():

		}
	}()

	return &Reconnector{}
}
