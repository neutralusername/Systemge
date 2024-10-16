package ListenerWebsocket

import (
	"errors"

	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *WebsocketListener) StartAcceptRoutine(maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, acceptHandler Tools.AcceptHandler[SystemgeConnection.SystemgeConnection[[]byte]]) error {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	if listener.acceptRoutine != nil {
		return errors.New("receptionHandler is already running")
	}

	listener.acceptRoutine = Tools.NewRoutine(
		func(stopChannel <-chan struct{}) {
			if client, err := listener.accept(stopChannel); err == nil {
				acceptHandler(client)
			}
		},
		maxConcurrentHandlers, delayNs, timeoutNs,
	)

	return listener.acceptRoutine.StartRoutine()
}

func (listener *WebsocketListener) StopAcceptRoutine(abortOngoingCalls bool) error {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	if listener.acceptRoutine == nil {
		return errors.New("receptionHandler is not running")
	}

	err := listener.acceptRoutine.StopRoutine(abortOngoingCalls)
	listener.acceptRoutine = nil

	return err
}

func (listener *WebsocketListener) IsAcceptRoutineRunning() bool {
	listener.acceptMutex.RLock()
	defer listener.acceptMutex.RUnlock()

	return listener.acceptRoutine != nil
}
