package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketClient) StartReadRoutine(delayNs int64, asyncHandling bool, readHandler Tools.ReadHandler[*WebsocketClient]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.readHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.readHandler = readHandler
	client.readRoutineStopChannel = make(chan struct{})
	client.readRoutineWaitGroup.Add(1)
	go client.readRoutine(delayNs, asyncHandling)

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

func (client *WebsocketClient) readRoutine(delayNs int64, asyncHandling bool) {
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

			if asyncHandling {
				client.readRoutineWaitGroup.Add(1)
				go func(bytes []byte) {
					defer client.readRoutineWaitGroup.Done()
					client.readHandler(bytes, client)
				}(bytes)
			} else {
				client.readHandler(bytes, client)
			}
		}
	}
}
