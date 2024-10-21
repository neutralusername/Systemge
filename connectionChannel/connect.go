package connectionChannel

import (
	"errors"
	"time"

	"github.com/neutralusername/systemge/systemge"
)

func EstablishConnection[D any](connectionChannel chan<- *ConnectionRequest[D], timeoutNs int64) (systemge.Connection[D], error) {
	connectionRequest := &ConnectionRequest[D]{
		SendToListener:      make(chan D),
		ReceiveFromListener: make(chan D),
	}
	var deadline <-chan time.Time
	if timeoutNs > 0 {
		deadline = time.After(time.Duration(timeoutNs))
	}
	select {
	case connectionChannel <- connectionRequest:
		return New(connectionRequest.ReceiveFromListener, connectionRequest.SendToListener), nil
	case <-deadline:
		return nil, errors.New("timeout")
	}
}
