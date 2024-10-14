package WebsocketConnection

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketConnection) StartReadRoutine(maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, readHandler Tools.ReadHandler[[]byte, *WebsocketConnection]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine != nil {
		return errors.New("receptionHandler is already running")
	}

	client.readRoutine = Tools.NewRoutine(func(<-chan struct{}) {
		if bytes, err := client.Read(); err == nil {
			readHandler(bytes, client)
		}
	}, maxConcurrentHandlers, delayNs, timeoutNs)

	return client.readRoutine.StartRoutine()
}

func (client *WebsocketConnection) StopReadRoutine() error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine == nil {
		return errors.New("receptionHandler is not running")
	}

	err := client.readRoutine.StopRoutine()
	client.readRoutine = nil

	return err
}

func (client *WebsocketConnection) IsReadRoutineRunning() bool {
	client.readMutex.RLock()
	defer client.readMutex.RUnlock()

	return client.readRoutine != nil
}
