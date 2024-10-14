package ChannelConnection

import (
	"errors"
	"time"
)

// panic if listener stops while waiting for connection
func EstablishConnection[T any](connectionChannel ConnectionChannel[T], timeoutNs int64) (*ChannelConnection[T], error) {
	connectionRequest := &ConnectionRequest[T]{
		SendToListener:      make(chan T),
		ReceiveFromListener: make(chan T),
		Response:            make(chan bool),
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
