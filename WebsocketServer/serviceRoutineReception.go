package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func (server *WebsocketServer) receptionRoutine(websocketConnection *WebsocketConnection) {
	defer func() {
		server.onEvent___(Event.NewInfoNoOption(
			Event.ReceptionRoutineFinished,
			"stopped websocketConnection reception routine",
			Event.Context{
				Event.Circumstance: Event.ReceptionRoutine,
				Event.SessionId:    websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
			},
		))
		websocketConnection.waitGroup.Done()
	}()

	if event := server.onEvent___(Event.NewInfo(
		Event.ReceptionRoutineBegins,
		"started websocketConnection reception routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ReceptionRoutine,
			Event.SessionId:    websocketConnection.GetId(),
			Event.Address:      websocketConnection.GetAddress(),
		},
	)); !event.IsInfo() {
		return
	}

	for err := server.receiveMessage(websocketConnection); err == nil; {
	}

}

func (server *WebsocketServer) receiveMessage(websocketConnection *WebsocketConnection) error {
	messageBytes, err := server.read(websocketConnection, Event.ReceptionRoutine)
	if err != nil {
		server.onEvent___(Event.NewWarningNoOption(
			Event.ReadMessageFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.ReceptionRoutine,
				Event.SessionId:    websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
			},
		))
		return err
	}
	server.websocketConnectionMessagesReceived.Add(1)
	server.websocketConnectionMessagesBytesReceived.Add(uint64(len(messageBytes)))

	if server.config.HandleMessageReceptionSequentially {
		if err := server.handleReception(websocketConnection, messageBytes, Event.Sequential); err != nil {
			if event := server.onEvent___(Event.NewInfo(
				Event.HandleReceptionFailed,
				err.Error(),
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.ReceptionRoutine,
					Event.Behaviour:    Event.Sequential,
					Event.SessionId:    websocketConnection.GetId(),
					Event.Address:      websocketConnection.GetAddress(),
				},
			)); !event.IsInfo() {
				websocketConnection.Close()
				return event.GetError()
			} else {
				if server.config.PropagateMessageHandlerErrors {
					server.write(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ReceptionRoutine)
				}
			}
		}
	} else {
		websocketConnection.waitGroup.Add(1)
		go func() {
			defer websocketConnection.waitGroup.Done()
			if err := server.handleReception(websocketConnection, messageBytes, Event.Concurrent); err != nil {
				if event := server.onEvent___(Event.NewInfo(
					Event.HandleReceptionFailed,
					err.Error(),
					Event.Cancel,
					Event.Cancel,
					Event.Continue,
					Event.Context{
						Event.Circumstance: Event.ReceptionRoutine,
						Event.Behaviour:    Event.Concurrent,
						Event.SessionId:    websocketConnection.GetId(),
						Event.Address:      websocketConnection.GetAddress(),
					},
				)); !event.IsInfo() {
					websocketConnection.Close()
				} else {
					if server.config.PropagateMessageHandlerErrors {
						server.write(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ReceptionRoutine)
					}
				}
			}
		}()
	}
	return nil
}

func (server *WebsocketServer) handleReception(websocketConnection *WebsocketConnection, messageBytes []byte, behaviour string) error {
	result, err := websocketConnection.pipeline.Process(messageBytes)

	if err != nil {
		server.onEvent___(Event.NewWarningNoOption(
			Event.PipelineFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.HandleReception,
				Event.Behaviour:    behaviour,
				Event.SessionId:    websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
				Event.Bytes:        string(messageBytes),
			},
		))
		return err
	}
	message := result.(*Message.Message)

	_, err = server.topicManager.HandleTopic(message.GetTopic(), websocketConnection, message)
	return err
}

func (server *WebsocketServer) validator(data any) error {
	message := data.(*Message.Message)
	if len(message.GetSyncToken()) != 0 {
		return errors.New("message contains sync token")
	}
	if len(message.GetTopic()) == 0 {
		return errors.New("message missing topic")
	}
	if maxTopicSize := server.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return errors.New("message topic exceeds maximum size")
	}
	if maxPayloadSize := server.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return errors.New("message payload exceeds maximum size")
	}
	return nil
}

func (server *WebsocketServer) deserializer(data any) (any, error) {
	deserializeData := data.(*deserializeData)
	return Message.Deserialize(deserializeData.bytes, deserializeData.origin)
}

/*

	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()

	if handler == nil {
		server.onEvent___(Event.NewWarningNoOption(
			Event.NoHandlerForTopic,
			"no websocketConnection message handler for topic",
			Event.Context{
				Event.Circumstance: Event.HandleReception,
				Event.Behaviour:    behaviour,
				Event.HandlerType:  Event.WebsocketConnection,
				Event.SessionId:    websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))
		return errors.New("no handler for topic")
	}

	if err := handler(websocketConnection, message); err != nil {
		server.onEvent___(Event.NewWarningNoOption(
			Event.HandlerFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.HandleReception,
				Event.Behaviour:    behaviour,
				Event.HandlerType:  Event.WebsocketConnection,
				Event.SessionId:    websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))
		return err
	}

	server.onEvent___(Event.NewInfoNoOption(
		Event.HandledReception,
		"handled websocketConnection message",
		Event.Context{
			Event.Circumstance: Event.HandleReception,
			Event.Behaviour:    behaviour,
			Event.HandlerType:  Event.WebsocketConnection,
			Event.SessionId:    websocketConnection.GetId(),
			Event.Address:      websocketConnection.GetAddress(),
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	))
	return nil
*/
