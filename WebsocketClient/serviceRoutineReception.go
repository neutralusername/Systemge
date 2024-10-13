package WebsocketClient

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tools"
)

func (client *WebsocketClient) SetReceptionHandler(receptionHandler Tools.ReadHandler[*WebsocketClient]) error {
	client.readMutex.Lock()
	defer client.readMutex.Unlock()

	if client.receptionHandler != nil {
		return errors.New("receptionHandler is already running")
	}

	client.receptionHandler = receptionHandler
	go client.receptionRoutine()

	return nil
}

func (client *WebsocketClient) RemoveReceptionHandler() error {

}

func (client *WebsocketClient) receptionRoutine() {
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

	handleReceptionWrapper := func(bytes []byte) {
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

	for {
		messageBytes, err := client.Read(client.config.ReadTimeoutMs)
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

		if client.config.HandleMessagesSequentially {
			handleReceptionWrapper(messageBytes)
		} else {
			client.waitGroup.Add(1)
			go func(websocketClient *WebsocketClient.WebsocketClient) {
				handleReceptionWrapper(messageBytes)
				client.waitGroup.Done()
			}(client)
		}
	}
}
