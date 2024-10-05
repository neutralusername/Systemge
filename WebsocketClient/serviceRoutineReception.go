package WebsocketClient

import (
	"errors"
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func (connection *WebsocketClient) Read() (*Message.Message, error) {
	if connection.eventHandler != nil {
		if event := connection.onEvent(Event.New(
			Event.ReadingMessage,
			Event.Context{
				Event.Circumstance: Event.MessageReceptionRoutine,
			},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			return nil, errors.New("connection closed")
		}
	}

	connection.websocketConn.SetReadDeadline(time.Now().Add(time.Duration(connection.config.ReadDeadlineMs) * time.Millisecond))
	_, messageBytes, err := connection.websocketConn.ReadMessage()
	if err != nil {
		if connection.eventHandler != nil {
			connection.onEvent(Event.New(
				Event.ReadMessageFailed,
				Event.Context{
					Event.Circumstance: Event.MessageReceptionRoutine,
					Event.Error:        err.Error(),
				},
				Event.Continue,
			))
		}
		connection.Close()
		return nil, err
	}
	connection.bytesReceived.Add(uint64(len(messageBytes)))

	if connection.eventHandler != nil {
		if event := connection.onEvent(Event.New(
			Event.ReadMessage,
			Event.Context{
				Event.Circumstance: Event.MessageReceptionRoutine,
				Event.Bytes:        string(messageBytes),
			},
			Event.Continue,
			Event.Cancel,
		)); event.GetAction() == Event.Cancel {
			connection.rejectedMessagesReceived.Add(1)
			return nil, errors.New("read canceled")
		}
	}
	connection.messagesReceived.Add(1)

	message, err := connection.handleMessageReception(messageBytes, Event.Sequential)
	if err != nil {
		if connection.eventHandler != nil {
			connection.onEvent(Event.New(
				Event.HandleReceptionFailed,
				Event.Context{
					Event.Circumstance: Event.MessageReceptionRoutine,
					Event.Behaviour:    Event.Sequential,
					Event.Error:        err.Error(),
				},
				Event.Continue,
			))
		}
		connection.write(Message.NewAsync("error", err.Error()).Serialize(), Event.MessageReceptionRoutine)
		return nil, err
	}
	return message, nil
}

func (connection *WebsocketClient) handleMessageReception(messageBytes []byte, behaviour string) (*Message.Message, error) {
	event := connection.onEvent(Event.NewInfo(
		Event.HandlingMessageReception,
		"handling message reception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.HandleMessageReception,
			Event.Behaviour:    behaviour,
		},
	))
	if !event.IsInfo() {
		connection.rejectedMessagesReceived.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return event.GetError()
	}

	if connection.byteRateLimiter != nil && !connection.byteRateLimiter.Consume(uint64(len(messageBytes))) {
		if event := connection.onEvent(Event.NewWarning(
			Event.RateLimited,
			"byte rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:    Event.HandleMessageReception,
				Event.Behaviour:       behaviour,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Messages,
			},
		)); !event.IsInfo() {
			connection.rejectedMessagesReceived.Add(1)
			connection.messageChannelSemaphore.ReleaseBlocking()
			return event.GetError()
		}
	}

	if connection.messageRateLimiter != nil && !connection.messageRateLimiter.Consume(1) {
		if event := connection.onEvent(Event.NewWarning(
			Event.RateLimited,
			"message rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:    Event.HandleMessageReception,
				Event.Behaviour:       behaviour,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Messages,
			},
		)); !event.IsInfo() {
			connection.rejectedMessagesReceived.Add(1)
			connection.messageChannelSemaphore.ReleaseBlocking()
			return event.GetError()
		}
	}

	message, err := Message.Deserialize(messageBytes, connection.GetName())
	if err != nil {
		connection.invalidMessagesReceived.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		connection.onEvent(Event.NewWarningNoOption(
			Event.DeserializingFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.HandleMessageReception,
				Event.Behaviour:    behaviour,
				Event.StructType:   Event.Message,
				Event.Bytes:        string(messageBytes),
			},
		))
		return err
	}

	if err := connection.validateMessage(message); err != nil {
		if event := connection.onEvent(Event.NewWarning(
			Event.InvalidMessage,
			err.Error(),
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.HandleMessageReception,
				Event.Behaviour:    behaviour,
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
				Event.SyncToken:    message.GetSyncToken(),
			},
		)); !event.IsInfo() {
			connection.invalidMessagesReceived.Add(1)
			connection.messageChannelSemaphore.ReleaseBlocking()
			return event.GetError()
		}
	}

	if event := connection.onEvent(Event.NewInfo(
		Event.SendingToChannel,
		"sending message to channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.HandleMessageReception,
			Event.Behaviour:    behaviour,
			Event.ChannelType:  Event.MessageChannel,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		connection.rejectedMessagesReceived.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return event.GetError()
	}
	connection.priorityTokenQueue.AddItem("", message, 0, 0)

	connection.onEvent(Event.NewInfoNoOption(
		Event.HandledMessageReception,
		"handled message reception",
		Event.Context{
			Event.Circumstance: Event.HandleMessageReception,
			Event.Behaviour:    behaviour,
			Event.Topic:        message.GetTopic(),
			Event.Payload:      message.GetPayload(),
			Event.SyncToken:    message.GetSyncToken(),
		},
	))

	return nil
}

func (connection *WebsocketClient) validateMessage(message *Message.Message) error {
	if len(message.GetSyncToken()) != 0 {
		return errors.New("message contains sync token")
	}
	if len(message.GetTopic()) == 0 {
		return errors.New("message missing topic")
	}
	if maxTopicSize := connection.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return errors.New("message topic exceeds maximum size")
	}
	if maxPayloadSize := connection.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return errors.New("message payload exceeds maximum size")
	}
	return nil
}
