package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func (server *WebsocketServer) messageReceptionRoutine(websocketConnection *WebsocketConnection) {
	defer func() {
		server.onEvent___(Event.NewInfoNoOption(
			Event.MessageReceptionRoutineFinished,
			"finished websocketConnection message reception routine",
			Event.Context{
				Event.Circumstance: Event.MessageReceptionRoutine,
				Event.SessionId:    websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
			},
		))
		websocketConnection.waitGroup.Done()
	}()

	if event := server.onEvent___(Event.NewInfo(
		Event.MessageReceptionRoutineBegins,
		"beginning websocketConnection message reception routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.MessageReceptionRoutine,
			Event.SessionId:    websocketConnection.GetId(),
			Event.Address:      websocketConnection.GetAddress(),
		},
	)); !event.IsInfo() {
		return
	}

	for {
		messageBytes, err := server.read(websocketConnection, Event.MessageReceptionRoutine)
		if err != nil {
			server.onEvent___(Event.NewWarningNoOption(
				Event.ReadMessageFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance: Event.MessageReceptionRoutine,
					Event.SessionId:    websocketConnection.GetId(),
					Event.Address:      websocketConnection.GetAddress(),
				},
			))
			break
		}
		server.websocketConnectionMessagesReceived.Add(1)
		server.websocketConnectionMessagesBytesReceived.Add(uint64(len(messageBytes)))

		if server.config.HandleMessageReceptionSequentially {
			if err := server.handleMessageReception(websocketConnection, messageBytes, Event.Sequential); err != nil {
				if event := server.onEvent___(Event.NewInfo(
					Event.HandleReceptionFailed,
					err.Error(),
					Event.Cancel,
					Event.Cancel,
					Event.Continue,
					Event.Context{
						Event.Circumstance: Event.MessageReceptionRoutine,
						Event.Behaviour:    Event.Sequential,
						Event.SessionId:    websocketConnection.GetId(),
						Event.Address:      websocketConnection.GetAddress(),
					},
				)); !event.IsInfo() {
					websocketConnection.Close()
				} else {
					if server.config.PropagateMessageHandlerErrors {
						server.write(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.MessageReceptionRoutine)
					}
				}
			}
		} else {
			websocketConnection.waitGroup.Add(1)
			go func() {
				defer websocketConnection.waitGroup.Done()
				if err := server.handleMessageReception(websocketConnection, messageBytes, Event.Concurrent); err != nil {
					if event := server.onEvent___(Event.NewInfo(
						Event.HandleReceptionFailed,
						err.Error(),
						Event.Cancel,
						Event.Cancel,
						Event.Continue,
						Event.Context{
							Event.Circumstance: Event.MessageReceptionRoutine,
							Event.Behaviour:    Event.Concurrent,
							Event.SessionId:    websocketConnection.GetId(),
							Event.Address:      websocketConnection.GetAddress(),
						},
					)); !event.IsInfo() {
						websocketConnection.Close()
					} else {
						if server.config.PropagateMessageHandlerErrors {
							server.write(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.MessageReceptionRoutine)
						}
					}
				}
			}()
		}
	}
}

func (server *WebsocketServer) handleMessageReception(websocketConnection *WebsocketConnection, messageBytes []byte, behaviour string) error {
	result, err := websocketConnection.receptionManager.HandleReception(messageBytes, websocketConnection.GetId())

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

func (server *WebsocketServer) validator(data any, args ...any) error {
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
func (server *WebsocketServer) deserializer(bytes []byte, args ...any) (any, error) {
	return Message.Deserialize(bytes, args[0].(string))
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
