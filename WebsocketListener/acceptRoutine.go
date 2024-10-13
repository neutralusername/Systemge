package WebsocketListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Tools"
	"github.com/neutralusername/Systemge/WebsocketClient"
)

func (client *WebsocketListener) StartReadRoutine(config *Config.WebsocketClient, acceptHandler Tools.AcceptHandler[*WebsocketClient.WebsocketClient]) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.acceptHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.acceptHandler = acceptHandler
	client.acceptRoutineStopChannel = make(chan struct{})
	client.acceptRoutineWaitGroup.Add(1)
	go client.acceptRoutine(config)

	return nil
}

func (listener *WebsocketListener) StopReadRoutine() error {
	listener.mutex.Lock()
	defer listener.mutex.Unlock()

	if listener.acceptHandler == nil {
		return errors.New("receptionHandler is not running")
	}

	close(listener.acceptRoutineStopChannel)
	//listener.websocketConn.SetReadDeadline(time.Now())
	listener.acceptRoutineWaitGroup.Wait()
	listener.acceptHandler = nil
	return nil
}

func (listener *WebsocketListener) acceptRoutine(config *Config.WebsocketClient) {
	defer listener.acceptRoutineWaitGroup.Done()
	for {
		select {
		case <-listener.acceptRoutineStopChannel:
			return
		default:
			client, err := listener.accept(config)
			if err != nil {
				continue
			}
			listener.acceptHandler(client)
		}
	}
}
