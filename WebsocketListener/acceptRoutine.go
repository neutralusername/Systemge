package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (listener *WebsocketListener) StartAcceptRoutine(maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, acceptHandler Tools.AcceptHandler[*WebsocketClient.WebsocketClient]) error {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	if listener.acceptRoutine != nil {
		return errors.New("receptionHandler is already running")
	}

	listener.acceptRoutine = Tools.NewRoutine(func() {
		client, err := listener.accept(listener.acceptRoutine.GetStopChannel())
		if err != nil {
			listener.ClientsFailed.Add(1)
			return
		}
		listener.ClientsAccepted.Add(1)
		acceptHandler(client)
	}, maxConcurrentHandlers, delayNs, timeoutNs)

	return listener.acceptRoutine.StartRoutine()
}

func (client *WebsocketListener) StopAcceptRoutine() error {
	client.acceptMutex.Lock()
	defer client.acceptMutex.Unlock()

	if client.acceptRoutine == nil {
		return errors.New("receptionHandler is not running")
	}

	err := client.acceptRoutine.StopRoutine()
	client.acceptRoutine = nil

	return err
}

func (listener *WebsocketListener) IsAcceptRoutineRunning() bool {
	listener.acceptMutex.RLock()
	defer listener.acceptMutex.RUnlock()

	return listener.acceptRoutine != nil
}
