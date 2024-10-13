package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketClient) StartReadRoutine(receptionHandler Tools.ReadHandler[*WebsocketClient]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.receptionHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.receptionHandler = receptionHandler
	go client.readRoutine()

	return nil
}

func (client *WebsocketClient) StopReadRoutine() error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.receptionHandler == nil {
		return errors.New("receptionHandler is not running")
	}

	// close(readRoutineChannel)
	client.websocketConn.SetReadDeadline(time.Now())
	// wg.Wait()
	client.receptionHandler = nil
	return nil
}

func (client *WebsocketClient) readRoutine() {
	for {
		bytes, err := client.read()
		if err != nil {
			continue
		}

		client.receptionHandler(bytes, client)
	}
}
