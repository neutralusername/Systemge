package ConnectionChannel

import (
	"errors"
	"time"
)

func (connection *ChannelConnection[T]) Write(messageBytes T, timeoutNs int64) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()

	connection.writeDeadlineChange = make(chan struct{})
	connection.SetWriteDeadline(timeoutNs)

	for {
		select {
		case connection.sendChannel <- messageBytes:
			connection.writeDeadline = nil
			connection.writeDeadlineChange = nil
			connection.MessagesSent.Add(1)
			return nil

		case <-connection.writeDeadline:
			connection.writeDeadline = nil
			connection.writeDeadlineChange = nil
			return errors.New("timeout")

		case <-connection.writeDeadlineChange:
			continue
		}
	}

}

func (connection *ChannelConnection[T]) SetWriteDeadline(timeoutNs int64) {
	writeDeadlineChange := connection.writeDeadlineChange
	if writeDeadlineChange == nil {
		return
	}

	if timeoutNs > 0 {
		connection.writeDeadline = time.After(time.Duration(timeoutNs) * time.Nanosecond)
	} else {
		connection.writeDeadline = nil
	}

	connection.writeDeadlineChange = make(chan struct{})
	close(writeDeadlineChange)

	return
}
