package listenerChannel

import (
	"errors"
	"time"

	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/systemge"
)

func EstablishConnection[D any](connChann chan<- *connectionChannel.ConnectionRequest[D], timeoutNs int64) (systemge.Connection[D], error) {
	connectionRequest := &connectionChannel.ConnectionRequest[D]{
		SendToListener:      make(chan D),
		ReceiveFromListener: make(chan D),
	}
	var deadline <-chan time.Time
	if timeoutNs > 0 {
		deadline = time.After(time.Duration(timeoutNs))
	}
	select {
	case connChann <- connectionRequest:
		return connectionChannel.New(connectionRequest.ReceiveFromListener, connectionRequest.SendToListener), nil
	case <-deadline:
		return nil, errors.New("timeout")
	}
}
