package listenerChannel

import (
	"errors"

	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/status"
	"github.com/neutralusername/systemge/tools"
)

func (listener *ChannelListener[T]) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status == status.Started {
		return errors.New("tcpSystemgeListener is already started")
	}
	listener.sessionId = tools.GenerateRandomString(constants.SessionIdLength, tools.ALPHA_NUMERIC)
	listener.status = status.Pending
	listener.stopChannel = make(chan struct{})

	listener.status = status.Started

	return nil
}
