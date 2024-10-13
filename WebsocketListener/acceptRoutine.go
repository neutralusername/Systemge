package WebsocketListener

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

// if acceptRoutine is started and all already acceptRoutineSemaphore are taken, it won't do anything until they free up
func (client *WebsocketListener) StartAcceptRoutine(acceptHandler Tools.AcceptHandler[*WebsocketClient.WebsocketClient]) error {
	client.acceptMutex.Lock()
	defer client.acceptMutex.Unlock()

	if client.status != Status.Started {
		return errors.New("listener not started")
	}

	if client.acceptHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.acceptHandler = acceptHandler
	client.acceptRoutineStopChannel = make(chan struct{})
	client.acceptRoutineWaitGroup.Add(1)
	go client.acceptRoutine()

	return nil
}

func (listener *WebsocketListener) StopAcceptRoutine() error {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	if listener.acceptHandler == nil {
		return errors.New("receptionHandler is not running")
	}

	close(listener.acceptRoutineStopChannel)
	listener.acceptRoutineWaitGroup.Wait()
	listener.acceptHandler = nil
	return nil
}

func (listener *WebsocketListener) IsAcceptRoutineRunning() bool {
	listener.acceptMutex.RLock()
	defer listener.acceptMutex.RUnlock()

	return listener.acceptHandler != nil
}

func (listener *WebsocketListener) acceptRoutine(delayNs int64) {
	defer listener.acceptRoutineWaitGroup.Done()
	for {
		if delayNs > 0 {
			time.Sleep(time.Duration(delayNs) * time.Nanosecond)
		}
		select {
		case <-listener.acceptRoutineStopChannel:
			return

		case <-listener.acceptRoutineSemaphore.GetChannel():
			listener.acceptRoutineWaitGroup.Add(1)
			go func() {
				defer func() {
					listener.acceptRoutineSemaphore.Signal(struct{}{})
					listener.acceptRoutineWaitGroup.Done()
				}()

				select {
				case <-listener.acceptRoutineStopChannel:
					return

				default:
					client, err := listener.accept(listener.acceptRoutineStopChannel)
					if err != nil {
						listener.ClientsFailed.Add(1)
						return
					}
					listener.ClientsAccepted.Add(1)
					listener.acceptHandler(client)
				}
			}()
		}
	}
}
