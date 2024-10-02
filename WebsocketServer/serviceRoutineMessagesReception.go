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
	if websocketConnection.messageRateLimiter != nil {
		if !websocketConnection.messageRateLimiter.Consume(1) {
			event := server.onEvent(Event.NewWarning(
				Event.RateLimited,
				"tcpSystemgeConnection message rate limited",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				Event.Context{
					Event.Circumstance:    Event.HandleReception,
					Event.Behaviour:       behaviour,
					Event.RateLimiterType: Event.TokenBucket,
					Event.TokenBucketType: Event.Messages,
				},
			))
			if event.IsWarning() {
				server.write(websocketConnection, Message.NewAsync("error", "message rate limited").Serialize(), Event.MessageReceptionRoutine)
				return
			}
			if event.IsError() {
				websocketConnection.Close()
				return
			}
		}
	}
	if websocketConnection.byteRateLimiter != nil {
		if !websocketConnection.byteRateLimiter.Consume(uint64(len(messageBytes))) {
			event := server.onEvent(Event.NewWarning(
				Event.RateLimited,
				"tcpSystemgeConnection byte rate limited",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				Event.Context{
					Event.Circumstance:    Event.HandleReception,
					Event.Behaviour:       behaviour,
					Event.RateLimiterType: Event.TokenBucket,
					Event.TokenBucketType: Event.Bytes,
				},
			))
			if event.IsWarning() {
				server.write(websocketConnection, Message.NewAsync("error", "byte rate limited").Serialize(), Event.MessageReceptionRoutine)
				return
			}
			if event.IsError() {
				websocketConnection.Close()
				return
			}
		}
	}
	message, err := Message.Deserialize(messageBytes, websocketConnection.GetId())
	if err != nil {
		event := server.onEvent(Event.NewWarning(
			Event.DeserializingFailed,
			err.Error(),
			Event.Cancel,
			Event.Skip,
			Event.Skip,
			Event.Context{
				Event.Circumstance: Event.HandleReception,
				Event.Behaviour:    behaviour,
				Event.StructType:   Event.Message,
				Event.Bytes:        string(messageBytes),
			},
		))
		if event.IsError() {
			websocketConnection.Close()
			return
		} else {
			server.write(websocketConnection, Message.NewAsync("error", "deserializing failed").Serialize(), Event.MessageReceptionRoutine)
			return
		}
	}

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

	_, err = server.topicManager.HandleTopic(message.GetTopic(), websocketConnection, message)
	if err != nil {
		if server.config.PropagateMessageHandlerErrors {
			server.write(websocketConnection, Message.NewAsync("error", err.Error()).Serialize(), Event.MessageReceptionRoutine)
		}
	}

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
