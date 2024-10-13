package WebsocketClient

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketClient) StartReadRoutine(delayNs int64, maxActiveHandlers uint32, readHandler Tools.ReadHandler[*WebsocketClient]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readRoutine != nil {
		return errors.New("receptionHandler is already running")
	}

	client.readRoutine = Tools.NewRoutine(func() {
		if bytes, err := client.Read(); err == nil {
			readHandler(bytes, client)

		}
	}, 1, 0, 0)

	return client.readRoutine.StartRoutine()
}
