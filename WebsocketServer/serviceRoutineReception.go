package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
)

func (server *WebsocketServer) receptionRoutine(websocketConnection *WebsocketConnection) {
	defer func() {
		server.onEvent(Event.NewInfoNoOption(
			Event.ReceptionRoutineFinished,
			"stopped websocketConnection reception routine",
			Event.Context{
				Event.Circumstance:  Event.ReceptionRoutine,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetAddress(),
			},
		))
		websocketConnection.waitGroup.Done()
	}()

	if event := server.onEvent(Event.NewInfo(
		Event.ReceptionRoutineStarted,
		"started websocketConnection reception routine",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.ReceptionRoutine,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetAddress(),
		},
	)); !event.IsInfo() {
		return
	}

	for err := server.receiveMessage(websocketConnection); err == nil; {
	}

}

func (server *WebsocketServer) receiveMessage(websocketConnection *WebsocketConnection) error {
	messageBytes, err := server.receive(websocketConnection, Event.ReceptionRoutine)
	if err != nil {
		return err
	}
	server.websocketConnectionMessagesReceived.Add(1)
	server.websocketConnectionMessagesBytesReceived.Add(uint64(len(messageBytes)))

	if server.config.HandleMessageReceptionSequentially {
		if err := server.handleReception(websocketConnection, messageBytes); err != nil {
			if event := server.onEvent(Event.NewInfo(
				Event.HandleReceptionFailed,
				err.Error(),
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.HandleReception,
					Event.ClientType:    Event.WebsocketConnection,
					Event.ClientId:      websocketConnection.GetId(),
					Event.ClientAddress: websocketConnection.GetAddress(),
				},
			)); !event.IsInfo() {
				websocketConnection.Close()
				return event.GetError()
			} else {
				if server.config.PropagateMessageHandlerErrors {
					server.send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ReceptionRoutine)
				}
			}
		}
	} else {
		websocketConnection.waitGroup.Add(1)
		go func() {
			defer websocketConnection.waitGroup.Done()
			err := server.handleReception(websocketConnection, messageBytes)
			if event := server.onEvent(Event.NewInfo(
				Event.HandleReceptionFailed,
				err.Error(),
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.HandleReception,
					Event.ClientType:    Event.WebsocketConnection,
					Event.ClientId:      websocketConnection.GetId(),
					Event.ClientAddress: websocketConnection.GetAddress(),
				},
			)); !event.IsInfo() {
				websocketConnection.Close()
			} else {
				if server.config.PropagateMessageHandlerErrors {
					server.send(websocketConnection, Message.NewAsync("error", event.Marshal()).Serialize(), Event.ReceptionRoutine)
				}
			}
		}()
	}
	return nil
}

func (server *WebsocketServer) handleReception(websocketConnection *WebsocketConnection, messageBytes []byte) error {
	event := server.onEvent(Event.NewInfo(
		Event.HandlingReception,
		"handling message reception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.HandlerType:   Event.WebsocketConnection,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetAddress(),
			Event.Bytes:         string(messageBytes),
		},
	))
	if !event.IsInfo() {
		return event.GetError()
	}

	if websocketConnection.rateLimiterBytes != nil && !websocketConnection.rateLimiterBytes.Consume(uint64(len(messageBytes))) {
		if event := server.onEvent(Event.NewWarning(
			Event.RateLimited,
			"websocketConnection byte rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:    Event.HandleReception,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Bytes,
				Event.ClientType:      Event.WebsocketConnection,
				Event.ClientId:        websocketConnection.GetId(),
				Event.ClientAddress:   websocketConnection.GetAddress(),
				Event.Bytes:           string(messageBytes),
			},
		)); !event.IsInfo() {
			return event.GetError()
		}
	}

	if websocketConnection.rateLimiterMsgs != nil && !websocketConnection.rateLimiterMsgs.Consume(1) {
		if event := server.onEvent(Event.NewWarning(
			Event.RateLimited,
			"websocketConnection message rate limited",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:    Event.HandleReception,
				Event.RateLimiterType: Event.TokenBucket,
				Event.TokenBucketType: Event.Messages,
				Event.ClientType:      Event.WebsocketConnection,
				Event.ClientId:        websocketConnection.GetId(),
				Event.ClientAddress:   websocketConnection.GetAddress(),
				Event.Bytes:           string(messageBytes),
			},
		)); !event.IsInfo() {
			return event.GetError()
		}
	}

	message, err := Message.Deserialize(messageBytes, websocketConnection.GetId())
	if err != nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.DeserializingFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.StructType:    Event.Message,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetAddress(),
				Event.Bytes:         string(messageBytes),
			},
		))
		return err
	}

	if err := server.validateMessage(message); err != nil {
		if event := server.onEvent(Event.NewWarning(
			Event.InvalidMessage,
			err.Error(),
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetAddress(),
				Event.Topic:         message.GetTopic(),
				Event.Payload:       message.GetPayload(),
				Event.SyncToken:     message.GetSyncToken(),
			},
		)); !event.IsInfo() {
			return event.GetError()
		}
	}

	if message.GetTopic() == Message.TOPIC_HEARTBEAT {
		server.onEvent(Event.NewInfoNoOption(
			Event.HeartbeatReceived,
			"received websocketConnection heartbeat",
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetAddress(),
			},
		))
		return errors.New("received heartbeat")
	}

	server.messageHandlerMutex.Lock()
	handler := server.messageHandlers[message.GetTopic()]
	server.messageHandlerMutex.Unlock()

	if handler == nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.NoHandlerForTopic,
			"no websocketConnection message handler for topic",
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.HandlerType:   Event.WebsocketConnection,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetAddress(),
				Event.Topic:         message.GetTopic(),
				Event.Payload:       message.GetPayload(),
				Event.SyncToken:     message.GetSyncToken(),
			},
		))
		return errors.New("no handler for topic")
	}

	if event := server.onEvent(Event.NewInfo(
		Event.HandlingMessage,
		"handling websocketConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.HandlerType:   Event.WebsocketConnection,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetAddress(),
			Event.Topic:         message.GetTopic(),
			Event.Payload:       message.GetPayload(),
			Event.SyncToken:     message.GetSyncToken(),
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if err := handler(websocketConnection, message); err != nil {
		server.onEvent(Event.NewWarningNoOption(
			Event.HandlerFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance:  Event.HandleReception,
				Event.HandlerType:   Event.WebsocketConnection,
				Event.ClientType:    Event.WebsocketConnection,
				Event.ClientId:      websocketConnection.GetId(),
				Event.ClientAddress: websocketConnection.GetAddress(),
				Event.Topic:         message.GetTopic(),
				Event.Payload:       message.GetPayload(),
				Event.SyncToken:     message.GetSyncToken(),
			},
		))
		return err
	}

	server.onEvent(Event.NewInfoNoOption(
		Event.HandledMessage,
		"handled websocketConnection message",
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.HandlerType:   Event.WebsocketConnection,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetAddress(),
			Event.Topic:         message.GetTopic(),
			Event.Payload:       message.GetPayload(),
			Event.SyncToken:     message.GetSyncToken(),
		},
	))

	server.onEvent(Event.NewInfoNoOption(
		Event.HandledReception,
		"handled websocketConnection message",
		Event.Context{
			Event.Circumstance:  Event.HandleReception,
			Event.HandlerType:   Event.WebsocketConnection,
			Event.ClientType:    Event.WebsocketConnection,
			Event.ClientId:      websocketConnection.GetId(),
			Event.ClientAddress: websocketConnection.GetAddress(),
			Event.Topic:         message.GetTopic(),
			Event.Payload:       message.GetPayload(),
			Event.SyncToken:     message.GetSyncToken(),
		},
	))
	return nil
}

func (server *WebsocketServer) validateMessage(message *Message.Message) error {
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
