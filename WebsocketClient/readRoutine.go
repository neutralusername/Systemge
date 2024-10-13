package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketClient) StartReadRoutine(readHandler Tools.ReadHandler[*WebsocketClient]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.readHandler = readHandler
	client.readRoutineStopChannel = make(chan struct{})
	client.readRoutineWaitGroup.Add(1)
	go client.readRoutine()

	return nil
}

func (client *WebsocketClient) StopReadRoutine() error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readHandler == nil {
		return errors.New("receptionHandler is not running")
	}

	close(client.readRoutineStopChannel)
	client.websocketConn.SetReadDeadline(time.Now())
	client.readRoutineWaitGroup.Wait()
	client.readHandler = nil
	return nil
}

func (client *WebsocketClient) IsReadRoutineRunning() bool {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	return client.readHandler != nil
}

func (client *WebsocketClient) readRoutine() {
	defer client.readRoutineWaitGroup.Done()
	for {
		select {
		case <-client.readRoutineStopChannel:
			return
		default:
			bytes, err := client.read()
			if err != nil {
				continue
			}
			client.readHandler(bytes, client) // side effect to note: if a sub-goroutine is started in here, the .Stop() method will not wait for it to finish as was previously the case
		}
	}
}
