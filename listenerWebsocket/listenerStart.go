package listenerWebsocket

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/status"
	"github.com/neutralusername/Systemge/tools"
)

func (listener *WebsocketListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status == status.Started {
		return errors.New("tcpSystemgeListener is already started")
	}
	listener.sessionId = tools.GenerateRandomString(Constants.SessionIdLength, tools.ALPHA_NUMERIC)
	listener.status = status.Pending
	listener.stopChannel = make(chan struct{})

	if err := listener.httpServer.Start(); err != nil {
		close(listener.stopChannel)
		listener.status = status.Stopped
		return err
	}

	listener.status = status.Started

	return nil
}
