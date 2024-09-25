package TcpSystemgeConnection

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpSystemgeConnection) receptionRoutine() {
	defer func() {
		connection.onEvent(Event.NewInfoNoOption(
			Event.ClientReceptionRoutineFinished,
			"stopped tcpSystemgeConnection message reception",
			Event.Context{
				Event.Circumstance:  Event.ClientReceptionRoutine,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetIp(),
			},
		))
		connection.waitGroup.Done()
	}()

	if event := connection.onEvent(Event.NewInfo(
		Event.ClientReceptionRoutineStarted,
		"started tcpSystemgeConnection message reception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.ClientReceptionRoutine,
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
			Event.ReceivingClientMessage,
			"receiving websocketConnection message",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:  Event.ClientReceptionRoutine,
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
					Event.ReceivingClientMessageFailed,
					err.Error(),
					Event.Context{
						Event.Circumstance:  Event.ClientReceptionRoutine,
						Event.ClientType:    Event.TcpSystemgeConnection,
						Event.ClientName:    connection.GetName(),
						Event.ClientAddress: connection.GetIp(),
					}),
				)
				connection.Close()
				return errors.New("connection closed")
			}
			if event := connection.onEvent(Event.NewInfo(
				Event.ReceivingClientMessageFailed,
				err.Error(),
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.ClientReceptionRoutine,
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
			Event.ReceivedClientMessage,
			"received tcpSystemgeConnection message",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:  Event.ClientReceptionRoutine,
				Event.ClientType:    Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetIp(),
				Event.Bytes:         string(messageBytes),
			}),
		); !event.IsInfo() {
			return event.GetError()
		}

		if connection.config.HandleMessageReceptionSequentially {
			if err := connection.handleReception(messageBytes); err != nil {
				connection.messageChannelSemaphore.ReleaseBlocking()
			}
		} else {
			go func() { // finally possible thanks to semaphore usage
				if err := connection.handleReception(messageBytes); err != nil {
					connection.messageChannelSemaphore.ReleaseBlocking()
				}
			}()
		}
		return nil
	}
}

func (connection *TcpSystemgeConnection) handleReception(messageBytes []byte) error {
	if connection.rateLimiterBytes != nil && !connection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		connection.byteRateLimiterExceeded.Add(1)
		return Event.New("byte rate limiter exceeded", nil)
	}
	if connection.rateLimiterMessages != nil && !connection.rateLimiterMessages.Consume(1) {
		connection.messageRateLimiterExceeded.Add(1)
		return Event.New("message rate limiter exceeded", nil)
	}
	message, err := Message.Deserialize(messageBytes, connection.GetName())
	if err != nil {
		connection.invalidMessagesReceived.Add(1)
		return Event.New("failed to deserialize message", err)
	}
	if err := connection.validateMessage(message); err != nil {
		connection.invalidMessagesReceived.Add(1)
		return Event.New("failed to validate message", err)
	}
	if message.IsResponse() {
		if err := connection.addSyncResponse(message); err != nil {
			connection.invalidSyncResponsesReceived.Add(1)
			return Event.New("failed to add sync response", err)
		}
		connection.validMessagesReceived.Add(1)
		connection.messageChannelSemaphore.ReleaseBlocking()
		return nil
	} else {
		connection.validMessagesReceived.Add(1)
		connection.messageChannel <- message
		if connection.infoLogger != nil {
			connection.infoLogger.Log("Added message \"" + Helpers.GetPointerId(message) + "\" to processing channel")
		}
		return nil
	}
}

func (connection *TcpSystemgeConnection) validateMessage(message *Message.Message) error {
	if maxSyncTokenSize := connection.config.MaxSyncTokenSize; maxSyncTokenSize > 0 && len(message.GetSyncToken()) > maxSyncTokenSize {
		return Event.New("Message sync token exceeds maximum size", nil)
	}
	if len(message.GetTopic()) == 0 {
		return Event.New("Message missing topic", nil)
	}
	if maxTopicSize := connection.config.MaxTopicSize; maxTopicSize > 0 && len(message.GetTopic()) > maxTopicSize {
		return Event.New("Message topic exceeds maximum size", nil)
	}
	if maxPayloadSize := connection.config.MaxPayloadSize; maxPayloadSize > 0 && len(message.GetPayload()) > maxPayloadSize {
		return Event.New("Message payload exceeds maximum size", nil)
	}
	return nil
}
