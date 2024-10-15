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

	connection.SetReadDeadline(timeoutNs)

	for {
		select {
		case item := <-connection.receiveChannel:
			connection.readDeadline = nil
			connection.MessagesReceived.Add(1)
			return item, nil

		case <-connection.readDeadline:
			connection.readDeadline = nil
			var nilValue T
			return nilValue, errors.New("timeout")

		case <-connection.readDeadlineChange:
			continue
		}
	}
}

func (connection *ChannelConnection[T]) SetReadDeadline(timeoutNs int64) {
	readDeadlineChange := connection.readDeadlineChange

	if timeoutNs > 0 {
		connection.readDeadline = time.After(time.Duration(timeoutNs) * time.Nanosecond)
	} else {
		connection.readDeadline = nil
	}

	connection.readDeadlineChange = make(chan struct{})
	close(readDeadlineChange)

	return
}
