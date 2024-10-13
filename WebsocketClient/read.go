package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketClient) Read() ([]byte, error) {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.receptionHandler != nil {
		return nil, errors.New("receptionHandler is already running")
	}

	return client.read()
}

func (client *WebsocketClient) read() ([]byte, error) {
	_, messageBytes, err := client.websocketConn.ReadMessage()
	if err != nil {
		return nil, err
	}
	client.BytesReceived.Add(uint64(len(messageBytes)))
	client.MessagesReceived.Add(1)
	return messageBytes, nil
}

// can be used to cancel an ongoing read operation
func (client *WebsocketClient) SetReadDeadline(timeoutMs uint64) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.receptionHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond))
	return nil
}

func (client *WebsocketClient) StartReadHandler(receptionHandler Tools.ReadHandler[*WebsocketClient]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.receptionHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.receptionHandler = receptionHandler
	go client.readRoutine()

	return nil
}

func (client *WebsocketClient) StopReadHandler() error {
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
	defer func() {
		if client.eventHandler != nil {
			client.eventHandler.Handle(Event.New(
				Event.ReceptionRoutineEnds,
				Event.Context{
					Event.Address: client.GetAddress(),
				},
				Event.Continue,
				Event.Cancel,
			))
		}
		client.waitGroup.Done()
	}()

	if client.eventHandler != nil {
		event := client.eventHandler.Handle(Event.New(
			Event.ReceptionRoutineBegins,
			Event.Context{
				Event.Address: client.GetAddress(),
			},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	for {
		bytes, err := client.read()
		if err != nil {
			if client.eventHandler != nil {
				event := client.eventHandler.Handle(Event.New(
					Event.ReadMessageFailed,
					Event.Context{
						Event.Address: client.GetAddress(),
						Event.Error:   err.Error(),
					},
					Event.Cancel,
					Event.Skip,
				))
				if event.GetAction() == Event.Skip {
					continue
				}
			}
			client.Close()
			break
		}

		if err := client.receptionHandler(bytes, client); err != nil {
			if client.eventHandler != nil {
				event := client.eventHandler.Handle(Event.New(
					Event.ReceptionHandlerFailed,
					Event.Context{
						Event.Address: client.GetAddress(),
						Event.Error:   err.Error(),
						Event.Bytes:   string(bytes),
					},
					Event.Skip,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					client.Close()
				}
			}
		}
	}
}
