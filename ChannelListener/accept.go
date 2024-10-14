package WebsocketListener

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/ChannelConnection.go"
)

func (listener *ChannelListener[T]) accept(cancel <-chan struct{}) (*ChannelConnection.ChannelConnection[T], error) {
	select {
	case <-listener.stopChannel:
		return nil, errors.New("listener stopped")

	case <-cancel:
		return nil, errors.New("accept canceled")

	case connectionRequest := <-listener.connectionChannel:
		return ChannelConnection.New(connectionRequest.SendToListener, connectionRequest.ReceiveFromListener), nil
	}
}

func (listener *ChannelListener[T]) Accept() (*ChannelConnection.ChannelConnection[T], error) {
	return listener.accept(make(chan struct{}))
}

func (listener *ChannelListener[T]) AcceptTimeout(timeoutMs uint32) (*ChannelConnection.ChannelConnection[T], error) {
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