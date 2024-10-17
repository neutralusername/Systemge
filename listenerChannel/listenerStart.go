package listenerChannel

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/status"
	"github.com/neutralusername/Systemge/tools"
)

func (listener *ChannelListener[T]) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status == status.Started {
		return errors.New("tcpSystemgeListener is already started")
	}
	listener.sessionId = tools.GenerateRandomString(Constants.SessionIdLength, tools.ALPHA_NUMERIC)
	listener.status = status.Pending
	listener.stopChannel = make(chan struct{})

	listener.status = status.Started

	return nil
}
