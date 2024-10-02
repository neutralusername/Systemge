package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/TopicManager"
)

func (server *WebsocketServer) messageReceptionRoutine(websocketConnection *WebsocketConnection) {
	defer func() {
		if server.eventHandler != nil {
			server.onEvent(Event.NewInfoNoOption(
				Event.MessageReceptionRoutineFinished,
				"finished message reception routine",
				Event.Context{
					Event.Circumstance: Event.MessageReceptionRoutine,
					Event.SessionId:    websocketConnection.GetId(),
					Event.Address:      websocketConnection.GetAddress(),
				},
			))
		}
		websocketConnection.waitGroup.Done()
	}()

	if server.eventHandler != nil {
		if event := server.onEvent(Event.NewInfo(
			Event.MessageReceptionRoutineBegins,
			"beginning message reception routine",
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
	}

	for {
		messageBytes, err := server.read(websocketConnection, Event.MessageReceptionRoutine)
		if err != nil {
			if server.eventHandler != nil {
				server.onEvent(Event.NewWarningNoOption(
					Event.ReadMessageFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance: Event.MessageReceptionRoutine,
						Event.SessionId:    websocketConnection.GetId(),
						Event.Address:      websocketConnection.GetAddress(),
					},
				))
			}
			break
		}
		server.websocketConnectionMessagesReceived.Add(1)
		server.websocketConnectionMessagesBytesReceived.Add(uint64(len(messageBytes)))

		if server.config.HandleMessageReceptionSequentially {
			server.handleMessageReception(websocketConnection, messageBytes, Event.Sequential)
		} else {
			websocketConnection.waitGroup.Add(1)
			go func() {
				defer websocketConnection.waitGroup.Done()
				server.handleMessageReception(websocketConnection, messageBytes, Event.Sequential)
			}()
		}
	}
}

func (server *WebsocketServer) handleMessageReception(websocketConnection *WebsocketConnection, messageBytes []byte, behaviour string) {
	result, err := websocketConnection.receptionManager.HandleReception(messageBytes, websocketConnection)
	if err != nil {
		if server.eventHandler != nil {
			server.onEvent(Event.NewWarningNoOption(
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
		}
		if server.config.PropagateMessageHandlerErrors {
			server.write(websocketConnection, Message.NewAsync("error", err.Error()).Serialize(), Event.MessageReceptionRoutine)
		}
		return err
	}
	message := result.(*Message.Message)

	_, err = server.topicManager.HandleTopic(message.GetTopic(), websocketConnection, message)
	if err != nil {
		if server.config.PropagateMessageHandlerErrors {
			server.write(websocketConnection, Message.NewAsync("error", err.Error()).Serialize(), Event.MessageReceptionRoutine)
		}
	}

	return err
}

func (server *WebsocketServer) toTopicHandler(handler WebsocketMessageHandler) TopicManager.TopicHandler {
	return func(args ...any) (any, error) {
		websocketConnection := args[0].(*WebsocketConnection)
		message := args[1].(*Message.Message)

		if event := server.onEvent(Event.NewInfo(
			Event.HandlingTopic,
			"handling websocketConnection message",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.HandleTopic,
				Event.SessionId:    websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		)); !event.IsInfo() {
			return nil, errors.New("event cancelled")
		}

		err := handler(websocketConnection, message)
		if err != nil {
			server.onEvent(Event.NewWarningNoOption(
				Event.TopicHandlerFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance: Event.HandleTopic,
					Event.SessionId:    websocketConnection.GetId(),
					Event.Address:      websocketConnection.GetAddress(),
					Event.Topic:        message.GetTopic(),
					Event.Payload:      message.GetPayload(),
					Event.SyncToken:    message.GetSyncToken(),
				},
			))
		}

		server.onEvent(Event.NewInfoNoOption(
			Event.HandledTopic,
			"handled websocketConnection message",
			Event.Context{
				Event.Circumstance: Event.HandleTopic,
				Event.SessionId:    websocketConnection.GetId(),
				Event.Address:      websocketConnection.GetAddress(),
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		))

		return nil, err
	}
}

func (server *WebsocketServer) preValidator(bytes []byte, args ...any) error {
	websocketConnection := args[0].(*WebsocketConnection)
	if websocketConnection.byteRateLimiter != nil {
		if !websocketConnection.byteRateLimiter.Consume(uint64(len(bytes))) {
			return errors.New("rate limit exceeded")
		}
	}
	if websocketConnection.messageRateLimiter != nil {
		if !websocketConnection.messageRateLimiter.Consume(1) {
			return errors.New("rate limit exceeded")
		}
	}
	return nil
}
func (server *WebsocketServer) deserializer(bytes []byte, args ...any) (any, error) {
	return Message.Deserialize(bytes, args[0].(*WebsocketConnection).GetId())
}
func (server *WebsocketServer) postValidator(data any, args ...any) error {
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
