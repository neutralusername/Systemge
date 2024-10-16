package ConnectionChannel

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Systemge"
)

func EstablishConnection[T any](connectionChannel chan<- *ConnectionRequest[T], timeoutNs int64) (Systemge.SystemgeConnection[T], error) {
	connectionRequest := &ConnectionRequest[T]{
		SendToListener:      make(chan T),
		ReceiveFromListener: make(chan T),
	}
	var deadline <-chan time.Time
	if timeoutNs > 0 {
		deadline = time.After(time.Duration(timeoutNs))
	}
	select {
	case connectionChannel <- connectionRequest:
		return New(connectionRequest.ReceiveFromListener, connectionRequest.SendToListener), nil
	case <-deadline:
		return nil, errors.New("connection channel is full")
	default:
		return nil, errors.New("connection channel is full")
	}
}
