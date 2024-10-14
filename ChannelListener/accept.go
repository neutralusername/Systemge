package WebsocketListener

import (
	"time"

	"github.com/neutralusername/Systemge/WebsocketConnection"
)

func (listener *ChannelListener[T]) accept(cancel <-chan struct{}) (*WebsocketConnection.WebsocketConnection, error) {

}

func (listener *ChannelListener[T]) Accept() (*WebsocketConnection.WebsocketConnection, error) {
	return listener.accept(make(chan struct{}))
}

func (listener *ChannelListener[T]) AcceptTimeout(timeoutMs uint32) (*WebsocketConnection.WebsocketConnection, error) {
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
