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
			Event.ReceptionRoutineFinished,
			"stopped tcpSystemgeConnection message reception",
			Event.Context{
				Event.Circumstance:  Event.ReceptionRoutine,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetIp(),
			},
		))
		connection.waitGroup.Done()
	}()

	if event := connection.onEvent(Event.NewInfo(
		Event.ReceptionRoutineStarted,
		"started tcpSystemgeConnection message reception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.ReceptionRoutine,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetIp(),
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
			Event.ReceivingMessage,
			"receiving tcpSystemgeConnection message",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:  Event.ReceptionRoutine,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetIp(),
			}),
		); !event.IsInfo() {
			return event.GetError()
		}
		messageBytes, bytesReceived, err := connection.messageReceiver.ReceiveNextMessage()
		connection.bytesReceived.Add(uint64(bytesReceived))
		if err != nil {
			if Tcp.IsConnectionClosed(err) {
				connection.onEvent(Event.NewWarningNoOption(
					Event.ReceivingMessageFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance:  Event.ReceptionRoutine,
						Event.ClientType:    Event.TcpSystemgeConnection,
						Event.ClientName:    connection.GetName(),
						Event.ClientAddress: connection.GetIp(),
					}),
				)
				connection.Close()
				return errors.New("connection closed")
			}
			if event := connection.onEvent(Event.NewInfo(
				Event.ReceivingMessageFailed,
				err.Error(),
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.ReceptionRoutine,
					Event.ClientType:    Event.TcpSystemgeConnection,
					Event.ClientName:    connection.GetName(),
					Event.ClientAddress: connection.GetIp(),
				}),
			); !event.IsInfo() {
				return err
			} else {
				return nil
			}
		}

		if event := connection.onEvent(Event.NewInfo(
			Event.ReceivedMessage,
			"received tcpSystemgeConnection message",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:  Event.ReceptionRoutine,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetIp(),
				Event.Bytes:         string(messageBytes),
			}),
		); !event.IsInfo() {
			connection.rejectedMessages.Add(1)
			return event.GetError()
		}
		connection.messagesReceived.Add(1)

		if connection.config.HandleMessageReceptionSequentially {
			if event := connection.handleReception(messageBytes); event != nil {
				if event.IsError() {
					connection.Close()
				}
			}
		} else {
			go func() {
				if event := connection.handleReception(messageBytes); event != nil {
					if event.IsError() {
						connection.Close()
					}
				}
			}()
		}
		return nil
	}
}

func (connection *TcpSystemgeConnection) handleReception(messageBytes []byte) *Event.Event {
	event := connection.onEvent(Event.NewInfo(
		Event.HandlingReception,
		"handling tcpSystemgeConnection message reception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetIp(),
		},
	))
	if !event.IsInfo() {
		connection.rejectedMessages.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return event
	}

	if connection.rateLimiterBytes != nil && !connection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		if event := connection.onEvent(Event.NewWarning(
			Event.RateLimited,
			"tcpSystemgeConnection byte rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:    Event.HandleReception,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Messages,
				Event.ClientType:      Event.TcpSystemgeConnection,
				Event.ClientName:      connection.GetName(),
				Event.ClientAddress:   connection.GetIp(),
			},
		)); !event.IsInfo() {
			connection.rejectedMessages.Add(1)
			connection.messageChannelSemaphore.ReleaseBlocking()
			return event
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
				Event.Circumstance:    Event.HandleReception,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Messages,
				Event.ClientType:      Event.TcpSystemgeConnection,
				Event.ClientName:      connection.GetName(),
				Event.ClientAddress:   connection.GetIp(),
			},
		)); !event.IsInfo() {
			connection.rejectedMessages.Add(1)
			connection.messageChannelSemaphore.ReleaseBlocking()
			return event
		}
	}

	message, err := Message.Deserialize(messageBytes, connection.GetName())
	if err != nil {
		connection.invalidMessagesReceived.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return connection.onEvent(Event.NewWarningNoOption(
			Event.DeserializingFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.StructType:    Event.Message,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetIp(),
				Event.Bytes:         string(messageBytes),
			},
		))
	}

	if err := connection.validateMessage(message); err != nil {
		if event := connection.onEvent(Event.NewWarning(
			Event.InvalidMessage,
			err.Error(),
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetIp(),
				Event.Topic:         message.GetTopic(),
				Event.Payload:       message.GetPayload(),
				Event.SyncToken:     message.GetSyncToken(),
			},
		)); !event.IsInfo() {
			connection.invalidMessagesReceived.Add(1)
			connection.messageChannelSemaphore.ReleaseBlocking()
			return event
		}
	}

	if event := connection.onEvent(Event.NewInfo(
		Event.SendingToChannel,
		"sending message to channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.ChannelType:   Event.MessageChannel,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetIp(),
			Event.Topic:         message.GetTopic(),
			Event.Payload:       message.GetPayload(),
			Event.SyncToken:     message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		connection.rejectedMessages.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return event
	}
	connection.messageChannel <- message

	return connection.onEvent(Event.NewInfo(
		Event.SentToChannel,
		"sent message to channel",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.ChannelType:   Event.MessageChannel,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetIp(),
			Event.Topic:         message.GetTopic(),
			Event.Payload:       message.GetPayload(),
			Event.SyncToken:     message.GetSyncToken(),
		},
	))
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

/*

if message.IsResponse() {
	event := connection.onEvent(Event.NewInfo(
		Event.SyncResponseReceived,
		"received sync response",
		Event.Cancel,
		Event.Skip,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.ClientType:    Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetIp(),
			Event.Topic:         message.GetTopic(),
			Event.Payload:       message.GetPayload(),
			Event.SyncToken:     message.GetSyncToken(),
		},
	))
	if event.IsError() {
		connection.invalidSyncResponsesReceived.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return event
	}
	if event.IsInfo() {
		if err := connection.addSyncResponse(message); err != nil {
			connection.invalidSyncResponsesReceived.Add(1)
			if event := connection.onEvent(Event.NewWarning(
				Event.InvalidSyncResponse,
				err.Error(),
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.HandleReception,
					Event.ClientType:    Event.TcpSystemgeConnection,
					Event.ClientName:    connection.GetName(),
					Event.ClientAddress: connection.GetIp(),
					Event.Topic:         message.GetTopic(),
					Event.Payload:       message.GetPayload(),
					Event.SyncToken:     message.GetSyncToken(),
				},
			)); !event.IsInfo() {
				return event
			}
		} else {
			if event := connection.onEvent(Event.NewInfo(
				Event.ValidSyncResponse,
				"valid sync response",
				Event.Cancel,
				Event.Cancel,
				Event.Cancel,
				Event.Context{
					Event.Circumstance:  Event.HandleReception,
					Event.ClientType:    Event.TcpSystemgeConnection,
					Event.ClientName:    connection.GetName(),
					Event.ClientAddress: connection.GetIp(),
					Event.Topic:         message.GetTopic(),
					Event.Payload:       message.GetPayload(),
					Event.SyncToken:     message.GetSyncToken(),
				},
			)); !event.IsInfo() {
				return event
			}
			connection.validMessagesReceived.Add(1)
			connection.messageChannelSemaphore.ReleaseBlocking()
			return nil
		}
	}
}

*/
