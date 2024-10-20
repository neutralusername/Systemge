package listenerWebsocket

import (
	"errors"

	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

func (listener *WebsocketListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status == status.Started {
		return errors.New("tcpSystemgeListener is already started")
	}
	listener.sessionId = tools.GenerateRandomString(constants.SessionIdLength, tools.ALPHA_NUMERIC)
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
