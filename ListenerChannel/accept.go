package ListenerChannel

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/ConnectionChannel"
)

func (listener *ChannelListener[T]) accept(cancel <-chan struct{}) (*ConnectionChannel.ChannelConnection[T], error) {
	select {
	case <-listener.stopChannel:
		listener.ClientsFailed.Add(1)
		return nil, errors.New("listener stopped")

	case <-cancel:
		listener.ClientsFailed.Add(1)
		return nil, errors.New("accept canceled")

	case connectionRequest := <-listener.connectionChannel:
		listener.ClientsAccepted.Add(1)
		return ConnectionChannel.New(connectionRequest.SendToListener, connectionRequest.ReceiveFromListener), nil
	}
}

func (listener *ChannelListener[T]) Accept() (*ConnectionChannel.ChannelConnection[T], error) {
	return listener.accept(make(chan struct{}))
}

func (listener *ChannelListener[T]) AcceptTimeout(timeoutMs uint32) (*ConnectionChannel.ChannelConnection[T], error) {
	var deadline <-chan time.Time = time.After(time.Duration(timeoutMs) * time.Millisecond)
	var cancel chan struct{} = make(chan struct{})
	go func() {
		select {
		case <-deadline:
			close(cancel)
		case <-cancel:
			close(cancel)
		}
	}()
	websocketConnection, err := listener.accept(cancel)
	close(cancel)
	return websocketConnection, err
}
