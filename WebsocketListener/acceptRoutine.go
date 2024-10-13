package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

// if acceptRoutine is started and all already acceptRoutineSemaphore are taken, it won't do anything until they free up
func (client *WebsocketListener) StartAcceptRoutine(acceptHandler Tools.AcceptHandler[*WebsocketClient.WebsocketClient]) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

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
	listener.mutex.Lock()
	defer listener.mutex.Unlock()

	return listener.stopAcceptRoutine()
}

func (listener *WebsocketListener) stopAcceptRoutine() error {
	if listener.acceptHandler == nil {
		return errors.New("receptionHandler is not running")
	}

	close(listener.acceptRoutineStopChannel)
	listener.acceptRoutineWaitGroup.Wait()
	listener.acceptHandler = nil
	return nil
}

func (listener *WebsocketListener) IsAcceptRoutineRunning() bool {
	listener.mutex.RLock()
	defer listener.mutex.RUnlock()

	return listener.acceptHandler != nil
}

func (listener *WebsocketListener) acceptRoutine() {
	defer listener.acceptRoutineWaitGroup.Done()
	for {
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
