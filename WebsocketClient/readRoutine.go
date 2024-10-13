package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketClient) StartReadRoutine(delayNs int64, readHandler Tools.ReadHandler[*WebsocketClient]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.readHandler = readHandler
	client.readRoutineStopChannel = make(chan struct{})
	client.readRoutineWaitGroup.Add(1)
	go client.readRoutine(delayNs)

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
	client.readMutex.RLock()
	defer client.readMutex.RUnlock()

	return client.readHandler != nil
}

func (client *WebsocketClient) readRoutine(delayNs int64) {
	defer client.readRoutineWaitGroup.Done()
	for {
		if delayNs > 0 {
			time.Sleep(time.Duration(delayNs) * time.Nanosecond)
		}
		select {
		case <-client.readRoutineStopChannel:
			return
		default:
			bytes, err := client.read()
			if err != nil {
				continue
			}
			client.readHandler(bytes, client) // side effect to note: if a sub-goroutine is started in here, the .Stop() method will not wait for it to finish as was previously the case with "async handling"
		}
	}
}
