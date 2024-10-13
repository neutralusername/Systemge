package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (client *WebsocketListener) StartAcceptRoutine(acceptHandler Tools.AcceptHandler[*WebsocketClient.WebsocketClient]) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

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

	if listener.acceptHandler == nil {
		return errors.New("receptionHandler is not running")
	}

	close(listener.acceptRoutineStopChannel)
	listener.acceptRoutineWaitGroup.Wait()
	listener.acceptHandler = nil
	return nil
}

func (listener *WebsocketListener) acceptRoutine() {
	defer listener.acceptRoutineWaitGroup.Done()
	for {
		select {
		case <-listener.acceptRoutineStopChannel:
			return

		case <-listener.acceptRoutineSemaphore.GetChannel():
			listener.waitgroup.Add(1)
			go func() {
				defer func() {
					listener.acceptRoutineSemaphore.Signal(struct{}{})
					listener.waitgroup.Done()
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
