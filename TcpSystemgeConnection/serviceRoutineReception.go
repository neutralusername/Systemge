package TcpSystemgeConnection

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpSystemgeConnection) receptionRoutine() {
	defer func() {
		connection.onEvent(Event.NewInfoNoOption(
			Event.MessageReceptionRoutineFinished,
			"stopped tcpSystemgeConnection message reception",
			Event.Context{
				Event.Circumstance: Event.MessageReceptionRoutine,
			},
		))
		connection.waitGroup.Done()
	}()

	if event := connection.onEvent(Event.NewInfo(
		Event.MessageReceptionRoutineBegins,
		"started tcpSystemgeConnection message reception",
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

func (connection *TcpSystemgeConnection) receiveMessage() error {
	select {
	case <-connection.closeChannel:
		return errors.New("connection closed")
	case <-connection.messageChannelSemaphore.GetChannel():
		if event := connection.onEvent(Event.NewInfo(
			Event.ReadingMessage,
			"receiving tcpSystemgeConnection message",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.MessageReceptionRoutine,
			}),
		); !event.IsInfo() {
			return event.GetError()
		}
		messageBytes, err := connection.messageReceiver.ReadNextMessage()
		if err != nil {
			if Tcp.IsConnectionClosed(err) {
				connection.Close()
				return errors.New("connection closed")
			}
			if event := connection.onEvent(Event.NewInfo(
				Event.ReadMessageFailed,
				err.Error(),
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance: Event.MessageReceptionRoutine,
				}),
			); !event.IsInfo() {
				return err
			} else {
				return nil
			}
		}

		if event := connection.onEvent(Event.NewInfo(
			Event.ReadMessage,
			"received tcpSystemgeConnection message",
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
			if err := connection.handleReception(messageBytes, Event.Sequential); err != nil {
				if event := connection.onEvent(Event.NewInfo(
					Event.HandleReceptionFailed,
					err.Error(),
					Event.Cancel,
					Event.Cancel,
					Event.Continue,
					Event.Context{
						Event.Circumstance: Event.HandleMessageReception,
						Event.Behaviour:    Event.Sequential,
					},
				)); !event.IsInfo() {
					connection.Close()
				} else {
					/* if server.config.PropagateMessageHandlerErrors {
						server.send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ReceptionRoutine)
					} */
				}
			}
		} else {
			go func() {
				if err := connection.handleReception(messageBytes, Event.Concurrent); err != nil {
					if event := connection.onEvent(Event.NewInfo(
						Event.HandleReceptionFailed,
						err.Error(),
						Event.Cancel,
						Event.Cancel,
						Event.Continue,
						Event.Context{
							Event.Circumstance: Event.HandleMessageReception,
							Event.Behaviour:    Event.Concurrent,
						},
					)); !event.IsInfo() {
						connection.Close()
					} else {
						/* if server.config.PropagateMessageHandlerErrors {
							server.send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ReceptionRoutine)
						} */
					}
				}
			}()
		}
		return nil
	}
}

func (connection *TcpSystemgeConnection) handleReception(messageBytes []byte, behaviour string) error {
	event := connection.onEvent(Event.NewInfo(
		Event.HandlingMessageReception,
		"handling tcpSystemgeConnection message reception",
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

	if connection.rateLimiterBytes != nil && !connection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		if event := connection.onEvent(Event.NewWarning(
			Event.RateLimited,
			"tcpSystemgeConnection byte rate limited",
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

	if connection.rateLimiterMessages != nil && !connection.rateLimiterMessages.Consume(1) {
		if event := connection.onEvent(Event.NewWarning(
			Event.RateLimited,
			"tcpSystemgeConnection message rate limited",
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

	if message.IsResponse() {
		connection.messageChannelSemaphore.ReleaseBlocking()
		if err := connection.addSyncResponse(message); err != nil {
			connection.invalidMessagesReceived.Add(1)
			return err
		} else {
			return nil
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
		"handled tcpSystemgeConnection message reception",
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

func (connection *TcpSystemgeConnection) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := connection.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncToken()) > maxSyncTokenSize {
		return errors.New("message sync token exceeds maximum size")
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
