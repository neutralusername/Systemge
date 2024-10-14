package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketConnection"
)

func (listener *ChannelListener[T]) StartAcceptRoutine(maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, acceptHandler Tools.AcceptHandler[*WebsocketConnection.WebsocketConnection]) error {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	if listener.acceptRoutine != nil {
		return errors.New("receptionHandler is already running")
	}

	listener.acceptRoutine = Tools.NewRoutine(func(<-chan struct{}) {
		if client, err := listener.accept(listener.acceptRoutine.GetStopChannel()); err == nil {
			listener.ClientsAccepted.Add(1)
			acceptHandler(client)
		}
		listener.ClientsFailed.Add(1)

	}, maxConcurrentHandlers, delayNs, timeoutNs)

	return listener.acceptRoutine.StartRoutine()
}

func (client *ChannelListener[T]) StopAcceptRoutine() error {
	client.acceptMutex.Lock()
	defer client.acceptMutex.Unlock()

	if client.acceptRoutine == nil {
		return errors.New("receptionHandler is not running")
	}

	err := client.acceptRoutine.StopRoutine()
	client.acceptRoutine = nil

	return err
}

func (listener *ChannelListener[T]) IsAcceptRoutineRunning() bool {
	listener.acceptMutex.RLock()
	defer listener.acceptMutex.RUnlock()

	return listener.acceptRoutine != nil
}
