package ListenerWebsocket

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *WebsocketListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status == Status.Started {
		return errors.New("tcpSystemgeListener is already started")
	}
	listener.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	listener.status = Status.Pending
	listener.stopChannel = make(chan struct{})

	if err := listener.httpServer.Start(); err != nil {
		close(listener.stopChannel)
		listener.status = Status.Stopped
		return err
	}

	listener.status = Status.Started

	return nil
}
