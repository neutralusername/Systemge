package ConnectionChannel

import (
	"errors"
	"time"
)

func (connection *ChannelConnection[T]) Read(timeoutNs int64) (T, error) {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if connection.readRoutine != nil {
		var nilValue T
		return nilValue, errors.New("receptionHandler is already running")
	}

	connection.readDeadlineChange = make(chan struct{})
	connection.SetReadDeadline(timeoutNs)

	for {
		select {
		case item := <-connection.receiveChannel:
			connection.readDeadline = nil
			connection.readDeadlineChange = nil
			return item, nil
		case <-connection.readDeadline:
			connection.readDeadline = nil
			connection.readDeadlineChange = nil
			var nilValue T
			return nilValue, errors.New("timeout")
		case <-connection.readDeadlineChange:
			continue
		}
	}
}

func (connection *ChannelConnection[T]) SetReadDeadline(timeoutNs int64) {
	readDeadlineChange := connection.readDeadlineChange
	if readDeadlineChange == nil {
		return
	}

	if timeoutNs > 0 {
		connection.readDeadline = time.After(time.Duration(timeoutNs) * time.Nanosecond)
	} else {
		connection.readDeadline = nil
	}

	connection.readDeadlineChange = make(chan struct{})
	close(readDeadlineChange)

	return
}
