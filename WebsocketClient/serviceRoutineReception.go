package WebsocketClient

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func (connection *WebsocketClient) receptionRoutine() {
	defer func() {
		connection.onEvent(Event.NewInfoNoOption(
			Event.MessageReceptionRoutineFinished,
			"stopped message reception",
			Event.Context{
				Event.Circumstance: Event.MessageReceptionRoutine,
			},
		))
		connection.waitGroup.Done()
	}()

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageReceptionRoutineBegins,
		"started message reception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.MessageReceptionRoutine,
		},
	)); !event.IsInfo() {
		return
	}

	for err := connection.receiveMessage(); err == nil; {
	}
}

func (connection *WebsocketClient) receiveMessage() error {
	select {
	case <-connection.closeChannel:
		return errors.New("connection closed")
	case <-connection.messageChannelSemaphore.GetChannel():
		if event := connection.onEvent(Event.NewInfo(
			Event.ReadingMessage,
			"receiving message",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.MessageReceptionRoutine,
			}),
		); !event.IsInfo() {
			return event.GetError()
		}
		_, messageBytes, err := connection.websocketConn.ReadMessage()
		if err != nil {
			connection.onEvent(Event.NewWarningNoOption(
				Event.ReadMessageFailed,
				err.Error(),
				Event.Context{
					Event.Circumstance: Event.MessageReceptionRoutine,
				}),
			)
			connection.Close()
			return errors.New("connection closed")
		}
		connection.bytesReceived.Add(uint64(len(messageBytes)))

		if event := connection.onEvent(Event.NewInfo(
			Event.ReadMessage,
			"received message",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.MessageReceptionRoutine,
				Event.Bytes:        string(messageBytes),
			}),
		); !event.IsInfo() {
			connection.rejectedMessagesReceived.Add(1)
			return event.GetError()
		}
		connection.messagesReceived.Add(1)

		if connection.config.HandleMessageReceptionSequentially {
			if err := connection.handleMessageReception(messageBytes, Event.Sequential); err != nil {
				if event := connection.onEvent(Event.NewInfo(
					Event.HandleReceptionFailed,
					err.Error(),
					Event.Cancel,
					Event.Cancel,
					Event.Continue,
					Event.Context{
						Event.Circumstance: Event.MessageReceptionRoutine,
						Event.Behaviour:    Event.Sequential,
					},
				)); !event.IsInfo() {
					connection.Close()
				} else {
					connection.write(Message.NewAsync("error", err.Error()).Serialize(), Event.MessageReceptionRoutine)
				}
			}
		} else {
			go func() {
				if err := connection.handleMessageReception(messageBytes, Event.Concurrent); err != nil {
					if event := connection.onEvent(Event.NewInfo(
						Event.HandleReceptionFailed,
						err.Error(),
						Event.Cancel,
						Event.Cancel,
						Event.Continue,
						Event.Context{
							Event.Circumstance: Event.MessageReceptionRoutine,
							Event.Behaviour:    Event.Concurrent,
						},
					)); !event.IsInfo() {
						connection.Close()
					} else {
						connection.write(Message.NewAsync("error", err.Error()).Serialize(), Event.MessageReceptionRoutine)
					}
				}
			}()
		}
		return nil
	}
}

func (connection *WebsocketClient) handleMessageReception(messageBytes []byte, behaviour string) error {
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
	connection.messageChannel <- message

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
