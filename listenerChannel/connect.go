package listenerChannel

import (
	"errors"
	"time"

	"github.com/neutralusername/systemge/connectionChannel"
	"github.com/neutralusername/systemge/systemge"
)

func Connect[T any](
	connChann chan<- *connectionChannel.ConnectionRequest[T],
	timeoutNs int64,
) (systemge.Connection[T], error) {

	connectionRequest := &connectionChannel.ConnectionRequest[T]{
		SendToListener:      make(chan T),
		ReceiveFromListener: make(chan T),
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
