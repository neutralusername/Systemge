package ConnectionWebsocket

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

func (connection *WebsocketConnection) StartReadRoutine(maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, readHandler Tools.ReadHandler[[]byte, *WebsocketConnection]) error {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if connection.readRoutine != nil {
		return errors.New("receptionHandler is already running")
	}

	connection.readRoutine = Tools.NewRoutine(func(<-chan struct{}) {
		if bytes, err := connection.Read(); err == nil {
			readHandler(bytes, connection)
		}
	}, maxConcurrentHandlers, delayNs, timeoutNs)

	return connection.readRoutine.StartRoutine()
}

func (connection *WebsocketConnection) StopReadRoutine(abortOngoingCalls bool) error {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if connection.readRoutine == nil {
		return errors.New("receptionHandler is not running")
	}

	err := connection.readRoutine.StopRoutine(abortOngoingCalls)
	connection.readRoutine = nil

	return err
}

func (connection *WebsocketConnection) IsReadRoutineRunning() bool {
	connection.readMutex.RLock()
	defer connection.readMutex.RUnlock()

	return connection.readRoutine != nil
}
