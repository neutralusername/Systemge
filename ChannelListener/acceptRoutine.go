package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/ChannelConnection.go"
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *ChannelListener[T]) StartAcceptRoutine(maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, acceptHandler Tools.AcceptHandler[*ChannelConnection.ChannelConnection[T]]) error {
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

func (listener *ChannelListener[T]) StopAcceptRoutine() error {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	if listener.acceptRoutine == nil {
		return errors.New("receptionHandler is not running")
	}

	err := listener.acceptRoutine.StopRoutine()
	listener.acceptRoutine = nil

	return err
}

func (listener *ChannelListener[T]) IsAcceptRoutineRunning() bool {
	listener.acceptMutex.RLock()
	defer listener.acceptMutex.RUnlock()

	return listener.acceptRoutine != nil
}