package WebsocketClient

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketClient) StartReadRoutine(maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, readHandler Tools.ReadHandler[*WebsocketClient]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine != nil {
		return errors.New("receptionHandler is already running")
	}

	client.readRoutine = Tools.NewRoutine(func() {
		if bytes, err := client.Read(); err == nil {
			readHandler(bytes, client)
		}
	}, maxConcurrentHandlers, delayNs, timeoutNs)

	return client.readRoutine.StartRoutine()
}

func (client *WebsocketClient) StopReadRoutine() error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine == nil {
		return errors.New("receptionHandler is not running")
	}

	err := client.readRoutine.StopRoutine()
	client.readRoutine = nil

	return err
}
